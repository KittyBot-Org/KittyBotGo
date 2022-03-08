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
			userID     snowflake.Snowflake
			botList    = types.BotList(params["bot_list"])
			multiplier = 1
			err        error
		)
		defer r.Body.Close()

		if b.Config.BotLists.Tokens[botList] != r.Header.Get("Authorization") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch botList {
		case types.TopGG:
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		case types.BotListSpace:
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		case types.BotsForDiscordCom:
			var v voteAddPayload2
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User

		case types.DiscordBotListCom:
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.ID

		case types.DiscordservicesNet:
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			b.Logger.Error("Error while handling bot list %s:", botList, err)
			return
		}
		if err = addVote(b, userID, botList, multiplier); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			b.Logger.Error("Error while handling bot list %s:", botList, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func addVote(b *types.Backend, userID snowflake.Snowflake, botList types.BotList, multiplier int) error {
	return b.RestServices.GuildService().AddMemberRole(b.Config.SupportGuildID, userID, b.Config.BotLists.VoterRoleID)
}
