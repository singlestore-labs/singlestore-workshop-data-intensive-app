package src

type Api struct {
	db *SingleStore
}

func NewApi(db *SingleStore) *Api {
	return &Api{db: db}
}
