package web

import (
	"embed"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"io/fs"
	"log"
	"net/http"
	"wholesale.local/orbit/internal/app"
	"wholesale.local/orbit/internal/cloud"
	"wholesale.local/orbit/internal/domain"
)

//go:embed static/*
var assets embed.FS

type API struct {
	Service *app.Service
	AWS     *cloud.AWS
}

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, e error) {
	respond(w, status, map[string]string{"error": e.Error()})
}
func (a *API) actor(r *http.Request) (domain.Actor, error) {
	var b []byte
	var x domain.Actor
	e := a.Service.DB.Pool.QueryRow(r.Context(), "SELECT body FROM actors WHERE id::text=$1", r.Header.Get("X-Actor-ID")).Scan(&b)
	if e != nil {
		return x, errors.New("利用者を選択してください")
	}
	e = json.Unmarshal(b, &x)
	return x, e
}
func (a *API) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		if e := a.Service.DB.Pool.Ping(r.Context()); e != nil {
			fail(w, 503, e)
			return
		}
		respond(w, 200, map[string]bool{"ok": true})
	})
	m.HandleFunc("GET /api/bootstrap", a.bootstrap)
	m.HandleFunc("GET /api/state", a.state)
	m.HandleFunc("POST /api/commands", a.command)
	m.HandleFunc("POST /api/commands/{id}/retry", a.retry)
	m.HandleFunc("GET /api/invoices/{id}", a.invoice)
	m.HandleFunc("GET /api/images/{id}", func(w http.ResponseWriter, r *http.Request) {
		b, e := a.AWS.Image(r.Context(), r.PathValue("id"))
		if e != nil {
			fail(w, 502, e)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=60")
		_, _ = w.Write(b)
	})
	m.HandleFunc("POST /api/maintenance", func(w http.ResponseWriter, r *http.Request) {
		actor, e := a.actor(r)
		if e != nil || actor.Role != "operator" {
			http.Error(w, "運用者のみ実行できます", 403)
			return
		}
		if e = a.AWS.Tick(r.Context()); e != nil {
			fail(w, 503, e)
			return
		}
		respond(w, 202, map[string]bool{"accepted": true})
	})
	sub, _ := fs.Sub(assets, "static")
	m.Handle("/", http.FileServer(http.FS(sub)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; object-src 'none'; frame-ancestors 'none'")
		m.ServeHTTP(w, r)
	})
}
func (a *API) bootstrap(w http.ResponseWriter, r *http.Request) {
	result := map[string]any{}
	for _, table := range []string{"actors", "products"} {
		rows, e := a.Service.DB.Pool.Query(r.Context(), "SELECT body FROM "+table+" ORDER BY body->>'name'")
		if e != nil {
			fail(w, 500, e)
			return
		}
		list := []json.RawMessage{}
		for rows.Next() {
			var b []byte
			if e = rows.Scan(&b); e != nil {
				break
			}
			list = append(list, json.RawMessage(b))
		}
		if e == nil {
			e = rows.Err()
		}
		rows.Close()
		if e != nil {
			fail(w, 500, e)
			return
		}
		result[table] = list
	}
	result["statuses"] = domain.StatusNames
	respond(w, 200, result)
}
func (a *API) state(w http.ResponseWriter, r *http.Request) {
	actor, e := a.actor(r)
	if e != nil {
		fail(w, 401, e)
		return
	}
	ctx := r.Context()
	rows, e := a.Service.DB.Pool.Query(ctx, "SELECT kind,id,version,body FROM projections ORDER BY kind,id")
	if e != nil {
		fail(w, 500, e)
		return
	}
	type view struct {
		Kind    string          `json:"kind"`
		ID      string          `json:"id"`
		Version int64           `json:"version"`
		Body    json.RawMessage `json:"body"`
	}
	list := []view{}
	positions := map[string]int64{}
	for rows.Next() {
		var x view
		if e = rows.Scan(&x.Kind, &x.ID, &x.Version, &x.Body); e != nil {
			break
		}
		positions[x.Kind+x.ID] = x.Version
		if x.Kind == "cart" || x.Kind == "loyalty" {
			if actor.Role != "operator" && x.ID != actor.ID {
				continue
			}
		}
		if x.Kind == "order" && actor.Role == "customer" {
			var o domain.Order
			_ = json.Unmarshal(x.Body, &o)
			if o.Snapshot().CustomerID != actor.ID {
				continue
			}
		}
		list = append(list, x)
	}
	if e == nil {
		e = rows.Err()
	}
	rows.Close()
	if e != nil {
		fail(w, 500, e)
		return
	}
	query := "SELECT id,actor_id,type,payload,status,result,error,attempts,created_at FROM commands WHERE actor_id=$1 OR $2='operator' ORDER BY created_at DESC LIMIT 50"
	rows, e = a.Service.DB.Pool.Query(ctx, query, actor.ID, actor.Role)
	if e != nil {
		fail(w, 500, e)
		return
	}
	cmds := []app.Command{}
	for rows.Next() {
		var c app.Command
		var p []byte
		if e = rows.Scan(&c.ID, &c.ActorID, &c.Type, &p, &c.Status, &c.Result, &c.Error, &c.Attempts, &c.CreatedAt); e != nil {
			break
		}
		_ = json.Unmarshal(p, &c.Payload)
		var result struct {
			Versions map[string]int64 `json:"versions"`
		}
		_ = json.Unmarshal(c.Result, &result)
		c.Projected = c.Status == "succeeded"
		for key, v := range result.Versions {
			if positions[key] < v {
				c.Projected = false
			}
		}
		cmds = append(cmds, c)
	}
	if e == nil {
		e = rows.Err()
	}
	rows.Close()
	if e != nil {
		fail(w, 500, e)
		return
	}
	rows, e = a.Service.DB.Pool.Query(ctx, "SELECT id,order_id,sent_at IS NOT NULL,last_error FROM invoices WHERE customer_id=$1 OR $2='operator' ORDER BY created_at DESC", actor.ID, actor.Role)
	if e != nil {
		fail(w, 500, e)
		return
	}
	invoices := []map[string]any{}
	for rows.Next() {
		var id, oid, err string
		var sent bool
		if e = rows.Scan(&id, &oid, &sent, &err); e != nil {
			break
		}
		invoices = append(invoices, map[string]any{"id": id, "orderId": oid, "sent": sent, "error": err})
	}
	if e == nil {
		e = rows.Err()
	}
	rows.Close()
	if e != nil {
		fail(w, 500, e)
		return
	}
	var now int64
	_ = a.Service.DB.Pool.QueryRow(ctx, "SELECT extract(epoch from clock_timestamp())::bigint").Scan(&now)
	respond(w, 200, map[string]any{"views": list, "commands": cmds, "invoices": invoices, "now": now})
}
func (a *API) command(w http.ResponseWriter, r *http.Request) {
	actor, e := a.actor(r)
	if e != nil {
		fail(w, 401, e)
		return
	}
	var c app.Command
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
	d.DisallowUnknownFields()
	if e = d.Decode(&c); e != nil {
		fail(w, 400, e)
		return
	}
	if c.Type == "cart.expire" {
		http.Error(w, "internal command", 403)
		return
	}
	c.ID = r.Header.Get("Idempotency-Key")
	c.ActorID = actor.ID
	status, e := a.Service.Submit(r.Context(), c)
	if e != nil {
		code := 400
		if errors.Is(e, app.ErrKeyReuse) {
			code = 409
		}
		fail(w, code, e)
		return
	}
	// Durable DB acceptance survives SQS unavailability. Scheduled outbox retries dispatch.
	if status == "queued" {
		if e = a.AWS.Send(r.Context(), c.ID); e != nil {
			log.Printf("dispatch pending %s: %v", c.ID, e)
		}
	}
	respond(w, 202, map[string]string{"id": c.ID, "status": status})
}
func (a *API) retry(w http.ResponseWriter, r *http.Request) {
	actor, e := a.actor(r)
	if e != nil || actor.Role != "operator" {
		http.Error(w, "運用者のみ実行できます", 403)
		return
	}
	tag, e := a.Service.DB.Pool.Exec(r.Context(), "UPDATE commands SET status='queued',attempts=0,error='',last_dispatched=NULL WHERE id::text=$1 AND status='dead'", r.PathValue("id"))
	if e != nil {
		fail(w, 500, e)
		return
	}
	if tag.RowsAffected() != 1 {
		http.Error(w, "再投入可能なCommandがありません", 409)
		return
	}
	_ = a.AWS.Send(r.Context(), r.PathValue("id"))
	respond(w, 202, map[string]bool{"accepted": true})
}
func (a *API) invoice(w http.ResponseWriter, r *http.Request) {
	// Browser download carries the selected demo actor as an explicit query parameter.
	if r.Header.Get("X-Actor-ID") == "" {
		r.Header.Set("X-Actor-ID", r.URL.Query().Get("actor"))
	}
	actor, e := a.actor(r)
	if e != nil {
		fail(w, 401, e)
		return
	}
	var pdf []byte
	e = a.Service.DB.Pool.QueryRow(r.Context(), "SELECT pdf FROM invoices WHERE id::text=$1 AND (customer_id=$2 OR $3='operator')", r.PathValue("id"), actor.ID, actor.Role).Scan(&pdf)
	if errors.Is(e, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if e != nil {
		fail(w, 500, e)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=invoice.pdf")
	_, _ = w.Write(pdf)
}
