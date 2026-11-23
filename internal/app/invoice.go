package app

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-pdf/fpdf"
	"github.com/jackc/pgx/v5"
	"net/smtp"
	"os"
	"time"
	"wholesale.local/orbit/internal/domain"
)

//go:embed fonts/BIZUDGothic-Regular.ttf
var invoiceFont []byte

func InvoicePDF(order domain.Order) ([]byte, error) {
	o := order.Snapshot()
	p := fpdf.New("P", "mm", "A4", "")
	p.AddUTF8FontFromBytes("jp", "", invoiceFont)
	p.AddPage()
	p.SetFont("jp", "", 20)
	p.CellFormat(180, 14, "請求書 / ORBIT Wholesale", "", 1, "L", false, 0, "")
	p.SetFont("jp", "", 10)
	lines := []string{"請求番号: INV-" + o.Number, "受注ID: " + o.ID, "発注日: " + o.CreatedAt.Format("2006-01-02"), o.Address.Company + " 御中", o.Address.Postal + " " + o.Address.Text, o.Address.Recipient + " 様 / " + o.Address.Phone, ""}
	for _, line := range lines {
		p.MultiCell(180, 7, line, "", "L", false)
	}
	for _, l := range o.Lines {
		p.MultiCell(180, 8, fmt.Sprintf("%s  %s\n  %d%s × %d円 = %d円", l.SKU, l.Name, l.Quantity, l.Unit, l.Price, l.Quantity*l.Price), "B", "L", false)
	}
	p.Ln(5)
	for _, line := range []string{fmt.Sprintf("税抜小計: %d円", o.Subtotal), fmt.Sprintf("割引 (%s / %d%%): -%d円", o.Coupon, o.DiscountPercent, o.Discount), fmt.Sprintf("割引後税抜: %d円", o.Net), fmt.Sprintf("消費税 10%%: %d円", o.Tax), fmt.Sprintf("ご請求額: %d円", o.Total), "擬似決済済み / 本書はローカル検証用です。"} {
		p.MultiCell(180, 8, line, "", "R", false)
	}
	var b bytes.Buffer
	e := p.Output(&b)
	return b.Bytes(), e
}
func (s *Service) SendInvoice(ctx context.Context) error {
	var id, customer string
	var pdf []byte
	// A bounded lease prevents concurrent senders; SMTP can still duplicate after acceptance + crash.
	e := s.DB.Pool.QueryRow(ctx, `UPDATE invoices SET lease_until=now()+interval '2 minutes',attempts=attempts+1 WHERE id=(SELECT id FROM invoices WHERE sent_at IS NULL AND (lease_until IS NULL OR lease_until<now()) ORDER BY created_at LIMIT 1) AND sent_at IS NULL AND (lease_until IS NULL OR lease_until<now()) RETURNING id,customer_id,pdf`).Scan(&id, &customer, &pdf)
	if errors.Is(e, pgx.ErrNoRows) {
		return nil
	}
	if e != nil {
		return e
	}
	var b []byte
	var a domain.Actor
	e = s.DB.Pool.QueryRow(ctx, "SELECT body FROM actors WHERE id=$1", customer).Scan(&b)
	if e != nil {
		return e
	}
	if e = json.Unmarshal(b, &a); e != nil {
		return e
	}
	boundary := "orbit-" + id
	msg := fmt.Sprintf("From: ORBIT <billing@orbit.example>\r\nTo: %s\r\nSubject: ORBIT Invoice %s\r\nMessage-ID: <%s@orbit.example>\r\nDate: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%q\r\n\r\n--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n配送が完了しました。請求書を添付します。\r\n\r\n--%s\r\nContent-Type: application/pdf\r\nContent-Disposition: attachment; filename=invoice.pdf\r\nContent-Transfer-Encoding: base64\r\n\r\n", a.Email, id, id, time.Now().Format(time.RFC1123Z), boundary, boundary, boundary)
	enc := base64.StdEncoding.EncodeToString(pdf)
	for len(enc) > 76 {
		msg += enc[:76] + "\r\n"
		enc = enc[76:]
	}
	msg += enc + "\r\n--" + boundary + "--\r\n"
	host := os.Getenv("SMTP_ADDR")
	if host == "" {
		host = "mailpit:1025"
	}
	e = smtp.SendMail(host, nil, "billing@orbit.example", []string{a.Email}, []byte(msg))
	if e != nil {
		_, _ = s.DB.Pool.Exec(ctx, "UPDATE invoices SET last_error=$2,lease_until=now()+interval '30 seconds' WHERE id=$1", id, e.Error())
		return e
	}
	_, e = s.DB.Pool.Exec(ctx, "UPDATE invoices SET sent_at=now(),last_error='',lease_until=NULL WHERE id=$1", id)
	return e
}
func (s *Service) Expire(ctx context.Context) error {
	rows, e := s.DB.Pool.Query(ctx, `SELECT id,body FROM aggregates WHERE kind='cart' AND (body->>'expires')::bigint<=extract(epoch from now())::bigint AND body->'items'<>'{}'::jsonb`)
	if e != nil {
		return e
	}
	var list []Command
	for rows.Next() {
		var id string
		var b []byte
		if e = rows.Scan(&id, &b); e != nil {
			rows.Close()
			return e
		}
		var c domain.Cart
		if e = json.Unmarshal(b, &c); e != nil {
			rows.Close()
			return e
		}
		list = append(list, Command{ID: stableID(fmt.Sprintf("expire/%s/%d", id, c.Snapshot().Expires)), ActorID: stableID("system"), Type: "cart.expire", Payload: Payload{CustomerID: id, Expires: c.Snapshot().Expires}})
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	rows, e = s.DB.Pool.Query(ctx, `SELECT id FROM aggregates WHERE kind='order' AND body->>'status'='0' AND (body->>'createdAt')::timestamptz<now()-interval '15 minutes'`)
	if e != nil {
		return e
	}
	for rows.Next() {
		var id string
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			return e
		}
		list = append(list, Command{ID: stableID("payment-timeout/" + id), ActorID: stableID("system"), Type: "order.cancel", Payload: Payload{OrderID: id, Note: "決済待ち15分超過"}})
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, c := range list {
		if _, e = s.Submit(ctx, c); e != nil {
			return e
		}
	}
	return nil
}
