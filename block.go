package goblet

type Block interface {
}

type HtmlGetBlock interface {
	Get(cx Context)
}

type HtmlPostBlock interface {
	Post(cx Context)
}

type RestNewBlock interface {
	New(cx Context)
}
