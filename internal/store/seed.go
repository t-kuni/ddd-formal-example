package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"wholesale.local/orbit/internal/domain"
)

func ID(name string) string {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("orbit.local/"+name)).String()
}
func (s *Store) Seed(ctx context.Context) error {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	actors := []domain.Actor{
		{ID: ID("customer-a"), Name: "青葉商事 / 佐藤", Role: "customer", Email: "sato@aoba.example", Address: domain.Address{Company: "青葉商事株式会社", Recipient: "佐藤 花子", Postal: "100-0005", Text: "東京都千代田区丸の内1-2-3 青葉ビル 3F", Phone: "03-0000-1001"}},
		{ID: ID("customer-b"), Name: "港フード / 鈴木", Role: "customer", Email: "suzuki@minato.example", Address: domain.Address{Company: "港フード株式会社", Recipient: "鈴木 太郎", Postal: "231-0001", Text: "神奈川県横浜市中区新港2-1-1", Phone: "045-000-2002"}},
		{ID: ID("warehouse-a"), Name: "東京倉庫 / 田中", Role: "warehouse", Email: "tanaka@orbit.example", Address: domain.Address{}},
		{ID: ID("warehouse-b"), Name: "東京倉庫 / 高橋", Role: "warehouse", Email: "takahashi@orbit.example", Address: domain.Address{}},
		{ID: ID("carrier"), Name: "東都配送 / 伊藤", Role: "carrier", Email: "ito@delivery.example", Address: domain.Address{}},
		{ID: ID("operator"), Name: "運用 / 山本", Role: "operator", Email: "ops@orbit.example", Address: domain.Address{}},
	}
	for _, a := range actors {
		b, _ := json.Marshal(a)
		if _, e = tx.Exec(ctx, "INSERT INTO actors VALUES($1,$2) ON CONFLICT DO NOTHING", a.ID, b); e != nil {
			return e
		}
	}
	type item struct {
		sku, name, category, unit, color string
		price, stock                     int64
	}
	list := []item{{"P-1001", "オーガニックコーヒー豆 1kg", "飲料", "袋", "#bb8056", 2400, 80}, {"P-1002", "天然水 500ml × 24本", "飲料", "ケース", "#5aa6bd", 1680, 120}, {"P-2001", "再生紙コピー用紙 A4 500枚", "事務用品", "冊", "#8b9b8a", 580, 160}, {"P-2002", "クラフト紙袋 M 100枚", "包装資材", "セット", "#c4a477", 1350, 60}, {"P-3001", "ニトリル手袋 M 100枚", "衛生用品", "箱", "#8c9cc1", 980, 90}, {"P-3002", "業務用洗剤 5L", "衛生用品", "本", "#78aa98", 1980, 45}}
	for _, x := range list {
		id := ID(x.sku)
		p := domain.Product{ID: id, SKU: x.sku, Name: x.name, Category: x.category, Unit: x.unit, Price: x.price, Image: "/api/images/" + id}
		b, _ := json.Marshal(p)
		svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="480" height="320" viewBox="0 0 480 320"><rect width="480" height="320" fill="#f3f1eb"/><ellipse cx="240" cy="270" rx="100" ry="14" fill="#ddd9d0"/><rect x="157" y="65" width="166" height="200" rx="14" fill="%s"/><rect x="178" y="130" width="124" height="74" rx="3" fill="#fffaf0"/><text x="240" y="162" text-anchor="middle" font-family="sans-serif" font-size="18" fill="#293b35">ORBIT</text><text x="240" y="186" text-anchor="middle" font-family="sans-serif" font-size="12" fill="#53665e">%s</text></svg>`, x.color, x.sku)
		if _, e = tx.Exec(ctx, "INSERT INTO products VALUES($1,$2,$3) ON CONFLICT DO NOTHING", id, b, []byte(svg)); e != nil {
			return e
		}
		quantity, _ := domain.NewQuantity(x.stock)
		stock, _ := domain.NewStock(quantity, domain.Quantity{})
		body, _ := json.Marshal(stock)
		if _, e = tx.Exec(ctx, "INSERT INTO aggregates VALUES('stock',$1,1,$2) ON CONFLICT DO NOTHING", id, body); e != nil {
			return e
		}
	}
	if _, e = tx.Exec(ctx, "INSERT INTO coupons(id,code,percent) VALUES($1,'WELCOME10',10),($2,'PARTNER5',5) ON CONFLICT DO NOTHING", ID("coupon/WELCOME10"), ID("coupon/PARTNER5")); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO projections SELECT * FROM aggregates ON CONFLICT DO NOTHING"); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
