package store

const ActorQuery = `SELECT jsonb_build_object('id',a.id,'name',a.name,'role',a.role,'email',a.email,
	'address',jsonb_build_object('company',COALESCE(c.company,''),'recipient',COALESCE(c.recipient,''),
		'postal',COALESCE(c.postal,''),'text',COALESCE(c.address,''),'phone',COALESCE(c.phone,'')))
	FROM actors a LEFT JOIN customer_addresses c ON c.customer_id=a.id`

const ProductQuery = `SELECT jsonb_build_object('id',id,'sku',sku,'name',name,'category',category,
	'unit',unit,'price',price,'image','/api/images/'||id) FROM products`
