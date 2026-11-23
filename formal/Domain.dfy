module Domain {
  predicate StockValid(h: int, r: int) { 0 <= r <= h <= 1000000000 }

  method Reserve(h: int, r: int, delta: int) returns (ok: bool, next: int)
    requires StockValid(h,r)
    ensures ok <==> 0 <= r + delta <= h
    ensures ok ==> next == r + delta && StockValid(h,next)
    ensures !ok ==> next == r
  {
    ok := 0 <= r + delta <= h;
    next := if ok then r + delta else r;
  }

  method Adjust(h: int, r: int, delta: int) returns (ok: bool, next: int)
    requires StockValid(h,r)
    ensures ok <==> r <= h + delta <= 1000000000
    ensures ok ==> next == h + delta && StockValid(next,r)
    ensures !ok ==> next == h
  {
    ok := r <= h + delta <= 1000000000;
    next := if ok then h + delta else h;
  }

  method Ship(h: int, r: int, qty: int) returns (ok: bool, nh: int, nr: int)
    requires StockValid(h,r)
    ensures ok <==> 0 < qty <= r
    ensures ok ==> nh == h-qty && nr == r-qty && StockValid(nh,nr)
    ensures ok ==> nh-nr == h-r
    ensures !ok ==> nh == h && nr == r
  {
    ok := 0 < qty <= r;
    nh := if ok then h-qty else h;
    nr := if ok then r-qty else r;
  }

  // 0 payment, 1 accepted, 2 picking, 3 blocked, 4 packed,
  // 5 shipping, 6 delivery_failed, 7 delivered, 8 cancelled,
  // 9 payment_failed, 10 returned.
  predicate Allowed(s: int, t: int) {
    (s == 0 && (t == 1 || t == 8 || t == 9)) ||
    (s == 1 && (t == 2 || t == 8)) ||
    (s == 2 && (t == 3 || t == 4 || t == 8)) ||
    (s == 3 && (t == 2 || t == 8)) ||
    (s == 4 && t == 5) ||
    (s == 5 && (t == 6 || t == 7)) ||
    (s == 6 && (t == 5 || t == 10))
  }
  method Transition(s: int, t: int) returns (ok: bool, next: int)
    requires 0 <= s <= 10
    ensures ok <==> Allowed(s,t)
    ensures next == (if ok then t else s)
    ensures 0 <= next <= 10
    ensures s in {7,8,9,10} ==> !ok && next == s
  { ok := Allowed(s,t); next := if ok then t else s; }

  method Touch(now: int) returns (expires: int)
    requires now >= 0
    ensures expires == now + 3600
    ensures expires > now
  { expires := now + 3600; }

  method Expired(now: int, expires: int) returns (expired: bool)
    ensures expired <==> now >= expires
  { expired := now >= expires; }

  method MayExpire(now: int, current: int, requested: int, hasItems: bool) returns (expire: bool)
    ensures expire <==> now >= current && current == requested && hasItems
    ensures current != requested ==> !expire
  { expire := now >= current && current == requested && hasItems; }

  method MayCheckout(now: int, expires: int, hasItems: bool) returns (allowed: bool)
    ensures allowed <==> now < expires && hasItems
    ensures now >= expires ==> !allowed
  { allowed := now < expires && hasItems; }

  method Price(subtotal: int, percent: int) returns (discount: int, net: int, tax: int, total: int, points: int)
    requires 0 <= subtotal <= 1000000000000
    requires 0 <= percent <= 100
    ensures discount == subtotal * percent / 100
    ensures net == subtotal - discount && 0 <= net <= subtotal
    ensures tax == net / 10 && total == net + tax
    ensures 0 <= total <= 1100000000000
    ensures points == net / 100 && points >= 0
  {
    discount := subtotal * percent / 100;
    assert 0 <= subtotal * percent <= subtotal * 100;
    net := subtotal - discount;
    tax := net / 10;
    total := net + tax;
    points := net / 100;
  }
}
