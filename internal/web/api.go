package web

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"

	"wholesale.local/orbit/api"
	"wholesale.local/orbit/internal/app"
	"wholesale.local/orbit/internal/cloud"
	"wholesale.local/orbit/internal/domain"
	"wholesale.local/orbit/internal/web/contract"
)

//go:embed static/*
var assets embed.FS

type API struct {
	Service *app.Service
	AWS     *cloud.AWS
}

func errorBody(err error) contract.Error {
	return contract.Error{Error: err.Error()}
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody(err))
}

func (a *API) actor(ctx context.Context, id string) (domain.Actor, error) {
	var body []byte
	var actor domain.Actor
	err := a.Service.DB.Pool.QueryRow(ctx, "SELECT body FROM actors WHERE id::text=$1", id).Scan(&body)
	if err != nil {
		return actor, errors.New("利用者を選択してください")
	}
	err = json.Unmarshal(body, &actor)
	return actor, err
}

func (a *API) Handler() http.Handler {
	spec, err := api.Load()
	if err != nil {
		panic(err)
	}
	badRequest := func(w http.ResponseWriter, _ *http.Request, err error) {
		writeError(w, http.StatusBadRequest, err)
	}
	strict := contract.NewStrictHandlerWithOptions(a, nil, contract.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: badRequest,
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeError(w, http.StatusInternalServerError, err)
		},
	})
	api := contract.HandlerWithOptions(strict, contract.StdHTTPServerOptions{
		ErrorHandlerFunc: badRequest,
	})
	validate := nethttpmiddleware.OapiRequestValidatorWithOptions(spec, &nethttpmiddleware.Options{
		// Bind directly from the original JSON so JPY and version int64s retain their precision.
		DiscardValidatedValue: true,
		ErrorHandler: func(w http.ResponseWriter, message string, status int) {
			writeError(w, status, errors.New(message))
		},
	})
	validatedAPI := validate(api)
	sub, _ := fs.Sub(assets, "static")
	static := http.NewServeMux()
	registerDocs(static)
	static.Handle("/", http.FileServer(http.FS(sub)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; object-src 'none'; frame-ancestors 'none'")
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			static.ServeHTTP(w, r)
			return
		}
		if r.Method == http.MethodHead {
			// ServeMux treats HEAD as GET. Match the same operation during validation.
			r = r.Clone(r.Context())
			r.Method = http.MethodGet
		}
		if r.Method == http.MethodPost && r.URL.Path == "/api/commands" {
			r.Body = http.MaxBytesReader(w, r.Body, 16384)
		}
		validatedAPI.ServeHTTP(w, r)
	})
}

func (a *API) GetHealth(ctx context.Context, _ contract.GetHealthRequestObject) (contract.GetHealthResponseObject, error) {
	if err := a.Service.DB.Pool.Ping(ctx); err != nil {
		return contract.GetHealth503JSONResponse(errorBody(err)), nil
	}
	return contract.GetHealth200JSONResponse{Ok: true}, nil
}

func loadMaster[T any](rows pgx.Rows) ([]T, error) {
	defer rows.Close()
	items := []T{}
	for rows.Next() {
		var body []byte
		if err := rows.Scan(&body); err != nil {
			return nil, err
		}
		var item T
		if err := json.Unmarshal(body, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (a *API) GetBootstrap(ctx context.Context, _ contract.GetBootstrapRequestObject) (contract.GetBootstrapResponseObject, error) {
	rows, err := a.Service.DB.Pool.Query(ctx, "SELECT body FROM actors ORDER BY body->>'name'")
	if err != nil {
		return contract.GetBootstrap500JSONResponse(errorBody(err)), nil
	}
	actors, err := loadMaster[contract.Actor](rows)
	if err != nil {
		return contract.GetBootstrap500JSONResponse(errorBody(err)), nil
	}
	rows, err = a.Service.DB.Pool.Query(ctx, "SELECT body FROM products ORDER BY body->>'name'")
	if err != nil {
		return contract.GetBootstrap500JSONResponse(errorBody(err)), nil
	}
	products, err := loadMaster[contract.Product](rows)
	if err != nil {
		return contract.GetBootstrap500JSONResponse(errorBody(err)), nil
	}
	return contract.GetBootstrap200JSONResponse{Actors: actors, Products: products, Statuses: domain.StatusNames}, nil
}

func (a *API) GetState(ctx context.Context, request contract.GetStateRequestObject) (contract.GetStateResponseObject, error) {
	actor, err := a.actor(ctx, value(request.Params.XActorID))
	if err != nil {
		return contract.GetState401JSONResponse(errorBody(err)), nil
	}
	views, positions, err := a.projections(ctx, actor)
	if err != nil {
		return contract.GetState500JSONResponse(errorBody(err)), nil
	}
	commands, err := a.commands(ctx, actor, positions)
	if err != nil {
		return contract.GetState500JSONResponse(errorBody(err)), nil
	}
	invoices, err := a.invoices(ctx, actor)
	if err != nil {
		return contract.GetState500JSONResponse(errorBody(err)), nil
	}
	var now int64
	_ = a.Service.DB.Pool.QueryRow(ctx, "SELECT extract(epoch from clock_timestamp())::bigint").Scan(&now)
	return contract.GetState200JSONResponse{Views: views, Commands: commands, Invoices: invoices, Now: now}, nil
}

func (a *API) projections(ctx context.Context, actor domain.Actor) ([]contract.Projection, map[string]int64, error) {
	rows, err := a.Service.DB.Pool.Query(ctx, "SELECT kind,id,version,body FROM projections ORDER BY kind,id")
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	list := []contract.Projection{}
	positions := map[string]int64{}
	for rows.Next() {
		var view struct {
			Kind    string          `json:"kind"`
			ID      string          `json:"id"`
			Version int64           `json:"version"`
			Body    json.RawMessage `json:"body"`
		}
		if err := rows.Scan(&view.Kind, &view.ID, &view.Version, &view.Body); err != nil {
			return nil, nil, err
		}
		positions[view.Kind+view.ID] = view.Version
		if (view.Kind == "cart" || view.Kind == "loyalty") && actor.Role != "operator" && view.ID != actor.ID {
			continue
		}
		if view.Kind == "order" && actor.Role == "customer" {
			var order domain.Order
			if err := json.Unmarshal(view.Body, &order); err != nil || order.Snapshot().CustomerID != actor.ID {
				continue
			}
		}
		data, err := json.Marshal(view)
		if err != nil {
			return nil, nil, err
		}
		var projection contract.Projection
		if err := projection.UnmarshalJSON(data); err != nil {
			return nil, nil, err
		}
		list = append(list, projection)
	}
	return list, positions, rows.Err()
}

func (a *API) commands(ctx context.Context, actor domain.Actor, positions map[string]int64) ([]contract.CommandRecord, error) {
	const query = "SELECT id,actor_id,type,payload,status,result,error,attempts,created_at FROM commands WHERE actor_id=$1 OR $2='operator' ORDER BY created_at DESC LIMIT 50"
	rows, err := a.Service.DB.Pool.Query(ctx, query, actor.ID, actor.Role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []contract.CommandRecord{}
	for rows.Next() {
		var c contract.CommandRecord
		var payload, result []byte
		if err := rows.Scan(&c.Id, &c.ActorId, &c.Type, &payload, &c.Status, &result, &c.Error, &c.Attempts, &c.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(payload, &c.Payload); err != nil {
			return nil, err
		}
		if len(result) != 0 {
			if err := json.Unmarshal(result, &c.Result); err != nil {
				return nil, err
			}
		}
		c.Projected = c.Status == contract.Succeeded
		if c.Result != nil && c.Result.Versions != nil {
			for key, version := range *c.Result.Versions {
				if positions[key] < version {
					c.Projected = false
					break
				}
			}
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (a *API) invoices(ctx context.Context, actor domain.Actor) ([]contract.Invoice, error) {
	const query = "SELECT id,order_id,sent_at IS NOT NULL,last_error FROM invoices WHERE customer_id=$1 OR $2='operator' ORDER BY created_at DESC"
	rows, err := a.Service.DB.Pool.Query(ctx, query, actor.ID, actor.Role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []contract.Invoice{}
	for rows.Next() {
		var invoice contract.Invoice
		if err := rows.Scan(&invoice.Id, &invoice.OrderId, &invoice.Sent, &invoice.Error); err != nil {
			return nil, err
		}
		list = append(list, invoice)
	}
	return list, rows.Err()
}

func value[T any](p *T) T {
	if p != nil {
		return *p
	}
	var zero T
	return zero
}

func commandPayload(dto contract.CommandPayload) app.Payload {
	payload := app.Payload{
		ProductID: value(dto.ProductId), Quantity: value(dto.Quantity), Delta: value(dto.Delta),
		OrderID: value(dto.OrderId), CustomerID: value(dto.CustomerId), ExpectedVersion: value(dto.ExpectedVersion),
		Expires: value(dto.Expires), Coupon: value(dto.Coupon), RequestedDate: value(dto.RequestedDate),
		Note: value(dto.Note), Success: value(dto.Success), BoxCount: value(dto.BoxCount), WeightGrams: value(dto.WeightGrams),
		Tracking: value(dto.Tracking), CarrierID: value(dto.CarrierId),
	}
	if dto.Address != nil {
		payload.Address = domain.Address{
			Company: value(dto.Address.Company), Recipient: value(dto.Address.Recipient), Postal: value(dto.Address.Postal),
			Text: value(dto.Address.Text), Phone: value(dto.Address.Phone),
		}
	}
	return payload
}

func (a *API) PostCommand(ctx context.Context, request contract.PostCommandRequestObject) (contract.PostCommandResponseObject, error) {
	actor, err := a.actor(ctx, value(request.Params.XActorID))
	if err != nil {
		return contract.PostCommand401JSONResponse(errorBody(err)), nil
	}
	c := app.Command{
		ID: request.Params.IdempotencyKey, ActorID: actor.ID, Type: string(request.Body.Type),
		Payload: commandPayload(request.Body.Payload),
	}
	status, err := a.Service.Submit(ctx, c)
	if errors.Is(err, app.ErrKeyReuse) {
		return contract.PostCommand409JSONResponse(errorBody(err)), nil
	}
	if err != nil {
		return contract.PostCommand400JSONResponse(errorBody(err)), nil
	}
	if status == "queued" {
		// Scheduled outbox processing retries dispatch after durable acceptance.
		if err := a.AWS.Send(ctx, c.ID); err != nil {
			log.Printf("dispatch pending %s: %v", c.ID, err)
		}
	}
	return contract.PostCommand202JSONResponse{Id: c.ID, Status: contract.CommandStatus(status)}, nil
}

func (a *API) RetryCommand(ctx context.Context, request contract.RetryCommandRequestObject) (contract.RetryCommandResponseObject, error) {
	actor, err := a.actor(ctx, value(request.Params.XActorID))
	if err != nil || actor.Role != "operator" {
		return contract.RetryCommand403JSONResponse{Error: "運用者のみ実行できます"}, nil
	}
	tag, err := a.Service.DB.Pool.Exec(ctx, "UPDATE commands SET status='queued',attempts=0,error='',last_dispatched=NULL WHERE id::text=$1 AND status='dead'", request.Id)
	if err != nil {
		return contract.RetryCommand500JSONResponse(errorBody(err)), nil
	}
	if tag.RowsAffected() != 1 {
		return contract.RetryCommand409JSONResponse{Error: "再投入可能なCommandがありません"}, nil
	}
	_ = a.AWS.Send(ctx, request.Id)
	return contract.RetryCommand202JSONResponse{Accepted: true}, nil
}

func (a *API) PostMaintenance(ctx context.Context, request contract.PostMaintenanceRequestObject) (contract.PostMaintenanceResponseObject, error) {
	actor, err := a.actor(ctx, value(request.Params.XActorID))
	if err != nil || actor.Role != "operator" {
		return contract.PostMaintenance403JSONResponse{Error: "運用者のみ実行できます"}, nil
	}
	if err := a.AWS.Tick(ctx); err != nil {
		return contract.PostMaintenance503JSONResponse(errorBody(err)), nil
	}
	return contract.PostMaintenance202JSONResponse{Accepted: true}, nil
}

func (a *API) GetImage(ctx context.Context, request contract.GetImageRequestObject) (contract.GetImageResponseObject, error) {
	body, err := a.AWS.Image(ctx, request.Id)
	if err != nil {
		return contract.GetImage502JSONResponse(errorBody(err)), nil
	}
	cache := "public, max-age=60"
	return contract.GetImage200ImagesvgXmlResponse{
		Body: bytes.NewReader(body), ContentLength: int64(len(body)),
		Headers: contract.GetImage200ResponseHeaders{CacheControl: &cache},
	}, nil
}

func (a *API) GetInvoice(ctx context.Context, request contract.GetInvoiceRequestObject) (contract.GetInvoiceResponseObject, error) {
	id := value(request.Params.XActorID)
	if id == "" {
		id = value(request.Params.Actor)
	}
	actor, err := a.actor(ctx, id)
	if err != nil {
		return contract.GetInvoice401JSONResponse(errorBody(err)), nil
	}
	var pdf []byte
	err = a.Service.DB.Pool.QueryRow(ctx, "SELECT pdf FROM invoices WHERE id::text=$1 AND (customer_id=$2 OR $3='operator')", request.Id, actor.ID, actor.Role).Scan(&pdf)
	if errors.Is(err, pgx.ErrNoRows) {
		return contract.GetInvoice404JSONResponse{Error: "請求書が見つかりません"}, nil
	}
	if err != nil {
		return contract.GetInvoice500JSONResponse(errorBody(err)), nil
	}
	disposition := "inline; filename=invoice.pdf"
	return contract.GetInvoice200ApplicationpdfResponse{
		Body: bytes.NewReader(pdf), ContentLength: int64(len(pdf)),
		Headers: contract.GetInvoice200ResponseHeaders{ContentDisposition: &disposition},
	}, nil
}
