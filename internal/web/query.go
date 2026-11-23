package web

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"wholesale.local/orbit/internal/web/contract"
)

func (a *API) GetCommands(ctx context.Context, request contract.GetCommandsRequestObject) (contract.GetCommandsResponseObject, error) {
	actor, err := a.actor(ctx, value(request.Params.XActorID))
	if err != nil {
		return contract.GetCommands401JSONResponse(errorBody(err)), nil
	}
	commands, err := a.commands(ctx, actor, request.Params.Id)
	if err != nil {
		return contract.GetCommands500JSONResponse(errorBody(err)), nil
	}
	return contract.GetCommands200JSONResponse(commands), nil
}

func (a *API) GetCatalog(ctx context.Context, request contract.GetCatalogRequestObject) (contract.GetCatalogResponseObject, error) {
	actor, err := a.actor(ctx, value(request.Params.XActorID))
	if err != nil {
		return contract.GetCatalog401JSONResponse(errorBody(err)), nil
	}
	result, err := a.catalog(ctx, actor.ID, actor.Role == "customer")
	if err != nil {
		return contract.GetCatalog500JSONResponse(errorBody(err)), nil
	}
	return contract.GetCatalog200JSONResponse(result), nil
}

func (a *API) catalog(ctx context.Context, id string, customer bool) (contract.CatalogState, error) {
	result := contract.CatalogState{Inventory: []contract.InventoryItem{}}
	tx, err := a.Service.DB.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	if err := tx.QueryRow(ctx, "SELECT floor(extract(epoch FROM now()))::bigint").Scan(&result.Now); err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, "SELECT product_id,version,on_hand,reserved FROM inventory ORDER BY product_id")
	if err != nil {
		return result, err
	}
	result.Inventory, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (contract.InventoryItem, error) {
		var item contract.InventoryItem
		err := row.Scan(&item.Id, &item.Version, &item.Body.OnHand, &item.Body.Reserved)
		return item, err
	})
	if err != nil {
		return result, err
	}
	if customer {
		var cart contract.CartState
		var items []byte
		err := tx.QueryRow(ctx, `SELECT version,expires_at,COALESCE(
			(SELECT jsonb_object_agg(product_id,quantity) FROM cart_items WHERE customer_id=c.customer_id),'{}')
			FROM carts c WHERE customer_id=$1`, id).Scan(&cart.Version, &cart.Body.Expires, &items)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
		case err != nil:
			return result, err
		default:
			cart.Body.CustomerId = id
			if err := json.Unmarshal(items, &cart.Body.Items); err != nil {
				return result, err
			}
			result.Cart = &cart
		}
	}
	return result, tx.Commit(ctx)
}

func (a *API) GetOrders(ctx context.Context, request contract.GetOrdersRequestObject) (contract.GetOrdersResponseObject, error) {
	actor, err := a.actor(ctx, value(request.Params.XActorID))
	if err != nil {
		return contract.GetOrders401JSONResponse(errorBody(err)), nil
	}
	var customer *string
	if actor.Role == "customer" {
		customer = &actor.ID
	}
	result, err := a.orderPage(ctx, customer, request.Params)
	if err != nil {
		return contract.GetOrders500JSONResponse(errorBody(err)), nil
	}
	return contract.GetOrders200JSONResponse(result), nil
}

const orderPageSize = 25

func (a *API) orderPage(ctx context.Context, customer *string, params contract.GetOrdersParams) (contract.OrderPage, error) {
	var result contract.OrderPage
	tx, err := a.Service.DB.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)

	const scope = `($1::uuid IS NULL OR o.customer_id=$1)
		AND ($2::text='' OR ($2='warehouse' AND o.status IN (1,2,3,4))
			OR ($2='delivery' AND o.status IN (4,5,6,7,10)))`
	queue := string(value(params.Queue))
	err = tx.QueryRow(ctx, `SELECT count(*),
		count(*) FILTER (WHERE o.status NOT IN (7,8,9,10)),
		count(*) FILTER (WHERE o.status=7),
		count(*) FILTER (WHERE o.status IN (3,6,9)),
		COALESCE((SELECT points FROM loyalty_accounts WHERE customer_id=$1),0)
		FROM orders o WHERE `+scope, customer, queue).Scan(
		&result.Metrics.Total, &result.Metrics.Active, &result.Metrics.Delivered, &result.Metrics.Exceptions, &result.Metrics.Points)
	if err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, `SELECT o.id,o.number,a.company,a.recipient,o.status,o.created_at,o.total,o.requested_date
		FROM orders o JOIN order_addresses a ON a.order_id=o.id
		WHERE `+scope+` AND ($3::int IS NULL OR o.status=$3)
		AND ($4::uuid IS NULL OR (o.created_at,o.id) <
			(SELECT created_at,id FROM orders WHERE id=$4 AND ($1::uuid IS NULL OR customer_id=$1)))
		ORDER BY o.created_at DESC,o.id DESC LIMIT $5`, customer, queue, params.Status, params.Before, orderPageSize+1)
	if err != nil {
		return result, err
	}
	result.Orders, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (contract.OrderSummary, error) {
		var item contract.OrderSummary
		err := row.Scan(&item.Id, &item.Number, &item.Company, &item.Recipient, &item.Status, &item.CreatedAt, &item.Total, &item.RequestedDate)
		return item, err
	})
	if err != nil {
		return result, err
	}
	if len(result.Orders) > orderPageSize {
		result.Orders = result.Orders[:orderPageSize]
		result.Next = &result.Orders[orderPageSize-1].Id
	}
	return result, tx.Commit(ctx)
}

func (a *API) GetOrder(ctx context.Context, request contract.GetOrderRequestObject) (contract.GetOrderResponseObject, error) {
	actor, err := a.actor(ctx, value(request.Params.XActorID))
	if err != nil {
		return contract.GetOrder401JSONResponse(errorBody(err)), nil
	}
	order, version, invoice, err := a.Service.DB.ReadOrderDetails(ctx, request.Id, actor.ID, actor.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return contract.GetOrder404JSONResponse(errorBody(errors.New("受注が見つかりません"))), nil
	}
	if err != nil {
		return contract.GetOrder500JSONResponse(errorBody(err)), nil
	}
	result := contract.OrderDetail{Version: version}
	body, err := json.Marshal(order)
	if err == nil {
		err = json.Unmarshal(body, &result.Order)
	}
	if err != nil {
		return contract.GetOrder500JSONResponse(errorBody(err)), nil
	}
	if invoice != nil {
		result.Invoice = &contract.Invoice{Id: invoice.ID, OrderId: request.Id, Sent: invoice.Sent, Error: invoice.Error}
	}
	return contract.GetOrder200JSONResponse(result), nil
}
