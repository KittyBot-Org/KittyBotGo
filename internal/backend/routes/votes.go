package routes

import (
	"encoding/json"
	"net/http"

	"github.com/DisgoOrg/snowflake"
	"github.com/KittyBot-Org/KittyBotGo/internal/backend/types"
	"github.com/gorilla/mux"
)

type voteAddPayload struct {
	User      user                `json:"user"`
	ID        snowflake.Snowflake `json:"id"`
	IsWeekend bool                `json:"isWeekend"`
}

type user struct {
	ID snowflake.Snowflake `json:"id"`
}

type voteAddPayload2 struct {
	User snowflake.Snowflake `json:"user"`
}

func VotesHandler(b *types.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		var (
			userID      snowflake.Snowflake
			botListName = params["bot_list"]
			botList     types.BotList
			multiplier  = 1
			err         error
		)
		defer r.Body.Close()

		if b.Config.BotLists.Tokens[botListName] != r.Header.Get("Authorization") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch botListName {
		case types.TopGG.Name:
			botList = types.TopGG
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		case types.BotListSpace.Name:
			botList = types.BotListSpace
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		case types.BotsForDiscordCom.Name:
			botList = types.BotsForDiscordCom
			var v voteAddPayload2
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User

		case types.DiscordBotListCom.Name:
			botList = types.DiscordBotListCom
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.ID

		case types.DiscordservicesNet.Name:
			botList = types.DiscordservicesNet
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			b.Logger.Error("Error while handling bot list %s:", botListName, err)
			return
		}
		if err = b.AddVote(userID, botList, multiplier); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			b.Logger.Error("Error while handling bot list %s:", botListName, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
