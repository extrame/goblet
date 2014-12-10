//goblet is a golang web framework
package goblet

//Block is a group of http request
type Block interface {
}

type HtmlGetBlock interface {
	Get(cx *Context)
}

type HtmlPostBlock interface {
	Post(cx *Context)
}

type RestCreateBlock interface {
	Create(cx *Context)
}

type RestNewBlock interface {
	New(cx *Context)
}

type RestReadManyBlock interface {
	ReadMany(cx *Context)
}

type RestReadBlock interface {
	Read(string, *Context)
}

type RestUpdateManyBlock interface {
	UpdateMany(*Context)
}

type RestDeleteManyBlock interface {
	DeleteMany(*Context)
}

type RestDeleteBlock interface {
	Delete(string, *Context)
}

type RestUpdateBlock interface {
	Update(string, *Context)
}

type RestEditBlock interface {
	Edit(string, *Context)
}
