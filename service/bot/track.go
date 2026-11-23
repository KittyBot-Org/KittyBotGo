package bot

import (
	"github.com/disgoorg/snowflake/v2"
)

type UserData struct {
	Requester snowflake.ID `json:"requester"`
}
