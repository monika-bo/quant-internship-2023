package dal

type UpdateOptions struct {
	Upsert bool
}

type UpdateOptionsFunc func(o *UpdateOptions)

var InsertIfNotFound UpdateOptionsFunc = func(o *UpdateOptions) {
	o.Upsert = true
}
