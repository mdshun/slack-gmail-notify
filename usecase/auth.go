package usecase

import (
	"github.com/mdshun/slack-gmail-notify/infra"
	"github.com/mdshun/slack-gmail-notify/repository"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

// AuthRequestInput is auth request param
type AuthRequestInput struct {
	Code  string
	State string
}

type authUsecaseImpl struct{}

// AuthUsecase is auth interface
type AuthUsecase interface {
	SlackAuth(ri *AuthRequestInput) error
}

// NewAuthUsecase will return auth usecase
func NewAuthUsecase() AuthUsecase {
	return &authUsecaseImpl{}
}

func (a *authUsecaseImpl) SlackAuth(ri *AuthRequestInput) error {
	or, err := slack.GetOAuthResponse(infra.Env.SlackClientID, infra.Env.SlackClientSecret, ri.Code, infra.Env.SlackRedirectedURL, infra.IsProduction())
	if err != nil {
		infra.Swarn(errWhileGetToken, ri, err)
		return errors.Wrap(err, errWhileGetToken)
	}

	tx := infra.RDB.Begin()

	team := &repository.Team{}

	infra.Sdebug("auth new team ", or)

	// check team instated?
	team.AccessToken = or.AccessToken
	team.Scope = or.Scope
	team.TeamName = or.TeamName
	team.TeamID = or.TeamID
	team.UserID = or.UserID
	team.BotAccessToken = or.Bot.BotAccessToken
	team.BotUserID = or.Bot.BotUserID

	infra.Sdebug("save team info ", team)

	teamRepo := repository.NewTeamRepository(tx)

	// check team was installed
	oldteam, err := teamRepo.FindByTeamID(team.TeamID)
	if err != nil {
		tx.Rollback()
		infra.Swarn(errWhileFindTeam, err)
		return errors.Wrap(err, errWhileFindTeam)
	}

	// have old team
	if oldteam != nil {
		oldteam.AccessToken = or.AccessToken
		oldteam.Scope = or.Scope
		oldteam.TeamName = or.TeamName
		oldteam.TeamID = or.TeamID
		oldteam.UserID = or.UserID
		oldteam.BotAccessToken = or.Bot.BotAccessToken
		oldteam.BotUserID = or.Bot.BotUserID

		_, err = teamRepo.Update(oldteam)

		if err != nil {
			tx.Rollback()
			infra.Swarn(errWhileSaveTeam, err)
			return errors.Wrap(err, errWhileSaveTeam)
		}

		tx.Commit()

		return nil
	}

	// is new team
	_, err = teamRepo.Add(team)
	if err != nil {
		tx.Rollback()
		infra.Swarn(errWhileSaveTeam, err)
		return errors.Wrap(err, errWhileSaveTeam)
	}

	// save user
	user := &repository.User{}
	user.UserID = team.UserID
	user.TeamID = team.TeamID

	infra.Sdebug("save user info ", user)

	userRepo := repository.NewUserRepository(tx)

	_, err = userRepo.Add(user)
	if err != nil {
		tx.Rollback()
		infra.Swarn(errWhileSaveUser, err)
		return errors.Wrap(err, errWhileSaveUser)
	}

	tx.Commit()

	return nil
}
