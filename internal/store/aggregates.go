package store

import (
	"github.com/jackc/pgx/v5/pgconn"
	"wholesale.local/orbit/internal/domain"
)

var aggregateQueries = map[domain.AggregateKind]string{
	domain.StockKind: `SELECT version,jsonb_build_object('onHand',on_hand,'reserved',reserved)
		FROM inventory WHERE product_id=$1`,
	domain.CartKind: `SELECT version,jsonb_build_object('customerId',customer_id,'expires',expires_at,
		'items',COALESCE((SELECT jsonb_object_agg(product_id,quantity) FROM cart_items WHERE customer_id=c.customer_id),'{}'))
		FROM carts c WHERE customer_id=$1`,
	domain.LoyaltyKind: `SELECT version,jsonb_build_object('points',points)
		FROM loyalty_accounts WHERE customer_id=$1`,
	domain.OrderKind: orderQuery,
}

const orderQuery = `SELECT o.version,jsonb_build_object(
		'id',o.id,'number',o.number,'customerId',o.customer_id,'status',o.status,
		'requestedDate',o.requested_date,'note',o.note,'coupon',o.coupon,'discountPercent',o.discount_percent,
		'subtotal',o.subtotal,'discount',o.discount,'net',o.net,'tax',o.tax,'total',o.total,'points',o.points,
		'workerId',COALESCE(o.worker_id::text,''),'carrierId',COALESCE(s.carrier_id::text,''),
		'tracking',COALESCE(s.tracking,''),'createdAt',o.created_at,
		'address',jsonb_build_object('company',a.company,'recipient',a.recipient,'postal',a.postal,'text',a.address,'phone',a.phone),
		'lines',COALESCE((SELECT jsonb_agg(jsonb_build_object('productId',product_id,'sku',sku,'name',name,'unit',unit,
			'quantity',quantity,'price',unit_price) ORDER BY position) FROM order_items WHERE order_id=o.id),'[]'),
		'boxes',COALESCE((SELECT jsonb_agg(jsonb_build_object('id',id,'label',label,'weightGrams',weight_grams)
			ORDER BY position) FROM shipment_boxes WHERE order_id=o.id),'[]'),
		'history',COALESCE((SELECT jsonb_agg(jsonb_build_object('at',occurred_at,'actor',actor_id,'from',status_from,
			'to',status_to,'note',note) ORDER BY sequence) FROM order_status_history WHERE order_id=o.id),'[]'))
		FROM orders o LEFT JOIN order_addresses a ON a.order_id=o.id LEFT JOIN shipments s ON s.order_id=o.id
		WHERE o.id=$1`

func saved(tag pgconn.CommandTag, err error) error {
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (u *Unit) save(id string, version int64, aggregate domain.Aggregate) error {
	switch value := aggregate.(type) {
	case *domain.Stock:
		return u.save(id, version, *value)
	case *domain.Cart:
		return u.save(id, version, *value)
	case *domain.Order:
		return u.save(id, version, *value)
	case *domain.Loyalty:
		return u.save(id, version, *value)
	case domain.Stock:
		x := value.Snapshot()
		if version == 0 {
			return saved(u.Tx.Exec(u.Ctx, `INSERT INTO inventory(product_id,version,on_hand,reserved)
				VALUES($1,1,$2,$3) ON CONFLICT DO NOTHING`, id, x.OnHand, x.Reserved))
		}
		return saved(u.Tx.Exec(u.Ctx, `UPDATE inventory SET version=version+1,on_hand=$3,reserved=$4
			WHERE product_id=$1 AND version=$2`, id, version, x.OnHand, x.Reserved))
	case domain.Cart:
		return u.saveCart(id, version, value.Snapshot())
	case domain.Order:
		return u.saveOrder(id, version, value.Snapshot())
	case domain.Loyalty:
		points := value.Balance().Value()
		if version == 0 {
			return saved(u.Tx.Exec(u.Ctx, `INSERT INTO loyalty_accounts(customer_id,version,points)
				VALUES($1,1,$2) ON CONFLICT DO NOTHING`, id, points))
		}
		return saved(u.Tx.Exec(u.Ctx, `UPDATE loyalty_accounts SET version=version+1,points=$3
			WHERE customer_id=$1 AND version=$2`, id, version, points))
	default:
		return domain.Invalid("保存できない集約です")
	}
}

func (u *Unit) saveCart(id string, version int64, cart domain.CartSnapshot) error {
	if id != cart.CustomerID {
		return domain.Invalid("カートの顧客IDが一致しません")
	}
	var err error
	if version == 0 {
		err = saved(u.Tx.Exec(u.Ctx, `INSERT INTO carts(customer_id,version,expires_at)
			VALUES($1,1,$2) ON CONFLICT DO NOTHING`, id, cart.Expires))
	} else {
		err = saved(u.Tx.Exec(u.Ctx, `UPDATE carts SET version=version+1,expires_at=$3
			WHERE customer_id=$1 AND version=$2`, id, version, cart.Expires))
	}
	if err != nil {
		return err
	}
	if _, err = u.Tx.Exec(u.Ctx, "DELETE FROM cart_items WHERE customer_id=$1", id); err != nil {
		return err
	}
	for product, quantity := range cart.Items {
		if _, err = u.Tx.Exec(u.Ctx, `INSERT INTO cart_items(customer_id,product_id,quantity)
			VALUES($1,$2,$3)`, id, product, quantity); err != nil {
			return err
		}
	}
	return nil
}

func (u *Unit) saveOrder(id string, version int64, order domain.OrderSnapshot) error {
	if id != order.ID {
		return domain.Invalid("受注IDが一致しません")
	}
	if version == 0 {
		if err := u.insertOrder(order); err != nil {
			return err
		}
	} else if err := saved(u.Tx.Exec(u.Ctx, `UPDATE orders SET version=version+1,status=$3,
		worker_id=NULLIF($4,'')::uuid WHERE id=$1 AND version=$2`, id, version, order.Status, order.WorkerID)); err != nil {
		return err
	}
	if len(order.Boxes) > 0 {
		if _, err := u.Tx.Exec(u.Ctx, `INSERT INTO shipments(order_id,carrier_id,tracking)
			VALUES($1,NULLIF($2,'')::uuid,$3) ON CONFLICT(order_id) DO UPDATE
			SET carrier_id=EXCLUDED.carrier_id,tracking=EXCLUDED.tracking`, id, order.CarrierID, order.Tracking); err != nil {
			return err
		}
		for i, box := range order.Boxes {
			if _, err := u.Tx.Exec(u.Ctx, `INSERT INTO shipment_boxes(id,order_id,position,label,weight_grams)
				VALUES($1,$2,$3,$4,$5) ON CONFLICT(order_id,position) DO NOTHING`, box.ID, id, i+1, box.Label, box.Weight); err != nil {
				return err
			}
		}
	}
	var recorded int
	if err := u.Tx.QueryRow(u.Ctx, `SELECT COALESCE(max(sequence),0) FROM order_status_history WHERE order_id=$1`, id).Scan(&recorded); err != nil {
		return err
	}
	if recorded > len(order.History) {
		return domain.Invalid("受注履歴が不足しています")
	}
	for i, event := range order.History[recorded:] {
		if _, err := u.Tx.Exec(u.Ctx, `INSERT INTO order_status_history
			(order_id,sequence,occurred_at,actor_id,status_from,status_to,note)
			VALUES($1,$2,$3,$4,$5,$6,$7)`, id, recorded+i+1, event.At, event.Actor, event.From, event.To, event.Note); err != nil {
			return err
		}
	}
	return nil
}

func (u *Unit) insertOrder(order domain.OrderSnapshot) error {
	if err := saved(u.Tx.Exec(u.Ctx, `INSERT INTO orders(id,version,number,customer_id,status,requested_date,
		note,coupon,discount_percent,subtotal,discount,net,tax,total,points,worker_id,created_at)
		VALUES($1,1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NULLIF($15,'')::uuid,$16)
		ON CONFLICT(id) DO NOTHING`, order.ID, order.Number, order.CustomerID, order.Status, order.RequestedDate,
		order.Note, order.Coupon, order.DiscountPercent, order.Subtotal, order.Discount, order.Net, order.Tax,
		order.Total, order.Points, order.WorkerID, order.CreatedAt)); err != nil {
		return err
	}
	address := order.Address
	if _, err := u.Tx.Exec(u.Ctx, `INSERT INTO order_addresses(order_id,company,recipient,postal,address,phone)
		VALUES($1,$2,$3,$4,$5,$6)`, order.ID, address.Company, address.Recipient, address.Postal, address.Text, address.Phone); err != nil {
		return err
	}
	for i, item := range order.Lines {
		if _, err := u.Tx.Exec(u.Ctx, `INSERT INTO order_items(order_id,position,product_id,sku,name,unit,quantity,unit_price)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, order.ID, i+1, item.ProductID, item.SKU, item.Name, item.Unit, item.Quantity, item.Price); err != nil {
			return err
		}
	}
	return nil
}
