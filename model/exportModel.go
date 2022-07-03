package model

type IcrmExport struct {
	Dbname string `json:"dbname"`
	Dirve  string `json:"dirve"`
	Sql    string `json:"sql"`
	Taskid string `json:"taskid"`
}
