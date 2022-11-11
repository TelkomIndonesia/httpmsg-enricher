package ecsx

import "github.com/telkomindonesia/httpmsg-enrichment/ecs"

type Document struct {
	ecs.Document
	HTTP *HTTP `json:"http"`

	CRS *CRS `json:"_crs"`
}
