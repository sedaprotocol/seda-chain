package types

type GetCountObj struct {
	Count int64 `json:"count"`
}

type GetCountResponse struct {
	// {"data":{"count":0}}
	Data *GetCountObj `json:"data"`
}
