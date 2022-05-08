package routes

import (
	"encoding/json"
	"net/http"

	"github.com/KittyBot-Org/KittyBotGo/internal/backend"

	"github.com/disgoorg/snowflake/v2"
	"github.com/gorilla/mux"
)

type voteAddPayload struct {
	User      user         `json:"user"`
	ID        snowflake.ID `json:"id"`
	IsWeekend bool         `json:"isWeekend"`
}

type user struct {
	ID snowflake.ID `json:"id"`
}

type voteAddPayload2 struct {
	User snowflake.ID `json:"user"`
}

func VotesHandler(b *backend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		var (
			userID      snowflake.ID
			botListName = params["bot_list"]
			botList     backend.BotList
			multiplier  = 1
			err         error
		)
		defer r.Body.Close()

		if b.Config.BotLists.Tokens[botListName] != r.Header.Get("Authorization") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch botListName {
		case backend.TopGG.Name:
			botList = backend.TopGG
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		case backend.BotListSpace.Name:
			botList = backend.BotListSpace
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		case backend.BotsForDiscordCom.Name:
			botList = backend.BotsForDiscordCom
			var v voteAddPayload2
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User

		case backend.DiscordBotListCom.Name:
			botList = backend.DiscordBotListCom
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.ID

		case backend.DiscordservicesNet.Name:
			botList = backend.DiscordservicesNet
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			b.Logger.Error("Error while handling kbot list %s:", botListName, err)
			return
		}
		if err = b.AddVote(userID, botList, multiplier); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			b.Logger.Error("Error while handling kbot list %s:", botListName, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
