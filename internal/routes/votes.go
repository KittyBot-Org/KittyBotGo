package routes

import (
	"encoding/json"
	"github.com/KittyBot-Org/KittyBotGo/internal/bend"
	"net/http"

	"github.com/disgoorg/snowflake"
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

func VotesHandler(b *bend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		var (
			userID      snowflake.Snowflake
			botListName = params["bot_list"]
			botList     bend.BotList
			multiplier  = 1
			err         error
		)
		defer r.Body.Close()

		if b.Config.BotLists.Tokens[botListName] != r.Header.Get("Authorization") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch botListName {
		case bend.TopGG.Name:
			botList = bend.TopGG
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		case bend.BotListSpace.Name:
			botList = bend.BotListSpace
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		case bend.BotsForDiscordCom.Name:
			botList = bend.BotsForDiscordCom
			var v voteAddPayload2
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User

		case bend.DiscordBotListCom.Name:
			botList = bend.DiscordBotListCom
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.ID

		case bend.DiscordservicesNet.Name:
			botList = bend.DiscordservicesNet
			var v voteAddPayload
			err = json.NewDecoder(r.Body).Decode(&v)
			userID = v.User.ID

		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			b.Logger.Error("Error while handling dbot list %s:", botListName, err)
			return
		}
		if err = b.AddVote(userID, botList, multiplier); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			b.Logger.Error("Error while handling dbot list %s:", botListName, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
