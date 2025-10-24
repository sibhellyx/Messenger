package authservice

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sibhellyx/Messenger/internal/models/entity"
	"github.com/sibhellyx/Messenger/internal/models/payload"
	"github.com/sibhellyx/Messenger/internal/models/request"
	"github.com/sibhellyx/Messenger/internal/models/response"
)

type RepositoryInterface interface {
	CreateSession(session entity.Session) error
	CreateUser(user entity.User) error
	DeleteSessionByUuid(uuid string) error
	FindJwtSessionByUuidAndRefreshToken(uuid string, refreshToken string) (*entity.Session, error)
	GetUserByCredentails(tgname string, password string) (*entity.User, error)
	UpdateSession(session entity.Session) error
	DeleteExpiredSessions(userId uint) error
	DeleteOldestSession(userId uint) error
	GetUserByTgname(tgname string) (*entity.User, error)
	CountActiveSessions(userId uint) (int64, error)
	ActivateUser(userId uint) error
}

type HasherInterface interface {
	Hash(password string) (string, error)
	HashRefreshToken(refreshToken string) string
	ComparePassword(hashedPassword, password string) bool
}

type TokenManagerInterface interface {
	NewJWT(payload payload.JwtPayload, ttl time.Duration) (string, error)
	NewRefreshToken() (string, error)
	Parse(accessToken string) (payload.JwtPayload, error)
}

type BotServiceInterface interface {
	GetLinkForFinishRegister(tgName string) string
}

type AuthService struct {
	repository RepositoryInterface

	hasher       HasherInterface
	tokenManager TokenManagerInterface
	bot          BotServiceInterface

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	activeSessions  int
}

func NewAuthService(
	repository RepositoryInterface,
	hasher HasherInterface,
	manager TokenManagerInterface,
	accessTokenTTL, refreshTokenTTL time.Duration,
	activeSessions int,
) *AuthService {
	return &AuthService{
		repository:      repository,
		hasher:          hasher,
		tokenManager:    manager,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		activeSessions:  activeSessions,
	}
}

func (s *AuthService) SetBotService(bot BotServiceInterface) {
	s.bot = bot
}

// register user service layer
func (s *AuthService) RegisterUser(r request.RegisterRequest) (string, error) {
	slog.Debug("service register started")
	err := r.Validate() //function return error if user not isValid
	if err != nil {
		slog.Error("error validating user input", "error", err.Error())
		return "", err
	}
	r.Password, err = s.hasher.Hash(r.Password) //hash password
	if err != nil {
		slog.Error("failed hash password", "error", err.Error())
		return "", err
	}

	// write repo
	err = s.repository.CreateUser(entity.User{
		Name:     r.Name,
		Surname:  r.Surname,
		Tgname:   r.Tgname,
		Password: r.Password,
		IsActive: false,
	})
	if err != nil {
		slog.Error("failed create user in repo", "error", err.Error())
		return "", err
	}
	// send to bot this user TgName and get link for register
	link := s.bot.GetLinkForFinishRegister(r.Tgname)

	slog.Debug("service register completed")
	return link, nil
}

// activate user account after start tg bot
func (s *AuthService) Activate(tgName string) error {
	slog.Debug("activating user", "tgName", tgName)
	user, err := s.repository.GetUserByTgname(tgName)
	if err != nil {
		slog.Error("failed get user", "error", err)
		return errors.New("failed get user")
	}
	err = s.repository.ActivateUser(user.ID)
	if err != nil {
		slog.Error("failed activate user", "error", err)
		return errors.New("failed activate user")
	}
	slog.Debug("activating user completed", "tgName", tgName)
	return nil
}

func (s *AuthService) SignInWithoutCode(user request.LoginRequest, params request.LoginParams) (response.Tokens, error) {
	slog.Debug("service login started")
	err := user.Validate()
	if err != nil {
		slog.Error("error validating user input", "error", err.Error())
		return response.Tokens{}, err
	}

	u, err := s.repository.GetUserByTgname(user.Tgname)
	if err != nil {
		slog.Error("failed to get user", "error", err.Error())
		return response.Tokens{}, errors.New("invalid credentials")
	}

	if !s.hasher.ComparePassword(u.Password, user.Password) {
		slog.Error("invalid password", "tgname", user.Tgname)
		// return false and error
		return response.Tokens{}, errors.New("invalid credentials")

	}

	return s.createSession(u.ID, params)
}

func (s *AuthService) SignIn(user request.LoginRequest, params request.LoginParams) (uint, error) {
	slog.Debug("service login started")
	err := user.Validate()
	if err != nil {
		slog.Error("error validating user input", "error", err.Error())
		return 0, err
	}

	u, err := s.repository.GetUserByTgname(user.Tgname)
	if err != nil {
		slog.Error("failed to get user", "error", err.Error())
		return 0, errors.New("invalid credentials")
	}

	if !s.hasher.ComparePassword(u.Password, user.Password) {
		slog.Error("invalid password", "tgname", user.Tgname)
		// return false and error
		return 0, errors.New("invalid credentials")

	}
	// return that now need input code from tg
	// generate code and send it in telegram to user
	// return userID and nil
	return u.ID, nil
}

func (s *AuthService) RefreshToken(payload payload.PayloadForRefresh, params request.LoginParams) (response.Tokens, error) {
	slog.Debug("service refresh token started")
	err := payload.Validate()
	if err != nil {
		slog.Error("error validating payload data from access token and refresh token", "error", err.Error())
		return response.Tokens{}, err
	}
	refreshTokenHash := s.hasher.HashRefreshToken(payload.RefreshToken)

	session, err := s.repository.FindJwtSessionByUuidAndRefreshToken(payload.Uuid, refreshTokenHash)
	if err != nil {
		slog.Error("failed founding session with this token token", "error", err.Error())
		return response.Tokens{}, errors.New("failed found session")
	}
	if session == nil {
		slog.Debug("session with this uuid and refresh token not found", "user_id", payload.UserId, "uuid", payload.Uuid, "refresh_token", payload.RefreshToken)
		return response.Tokens{}, errors.New("this session not found or this refresh token was issued separately")
	}
	if session.UserAgent != params.UserAgent {
		err := s.repository.DeleteSessionByUuid(payload.Uuid)
		if err != nil {
			slog.Error("failed deleted session, different user agents", "error", err.Error())
			return response.Tokens{}, errors.New("failed delete session, different user agents")
		}
		errMsg := "error refresh token from another user agent"
		slog.Error("different user agents", "error", errMsg)
		return response.Tokens{}, errors.New(errMsg)
	}
	return s.generateSessionAndSave(*session, false)
}

func (s *AuthService) createSession(userId uint, params request.LoginParams) (response.Tokens, error) {
	slog.Debug("creating session started")

	if err := s.repository.DeleteExpiredSessions(userId); err != nil {
		slog.Warn("failed to cleanup expired sessions", "error", err)
	}

	activeCount, err := s.repository.CountActiveSessions(userId)
	if err != nil {
		return response.Tokens{}, fmt.Errorf("failed to count active sessions: %w", err)
	}

	if activeCount >= int64(s.activeSessions) {
		if err := s.repository.DeleteOldestSession(userId); err != nil {
			slog.Warn("failed to delete oldest session", "error", err)
		}
	}

	session := entity.Session{
		UserID:    userId,
		UserAgent: params.UserAgent,
		LastIP:    params.LastIp,
	}

	return s.generateSessionAndSave(session, true)
}

func (s *AuthService) generateSessionAndSave(session entity.Session, isNewSession bool) (response.Tokens, error) {
	slog.Debug("generation session", "usesr_id", session.UserID)

	var (
		res response.Tokens
		err error
	)

	// creating tokens
	uid := uuid.New()
	payload := payload.JwtPayload{
		UserId: strconv.Itoa(int(session.UserID)),
		Uuid:   uid.String(),
	}

	slog.Debug("creating access token")
	res.AccessToken, err = s.tokenManager.NewJWT(payload, s.accessTokenTTL)
	if err != nil {
		slog.Error("failed create access token", "error", err)
		return response.Tokens{}, errors.New("failed create access token")
	}

	slog.Debug("creating refresh token")
	res.RefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		slog.Error("failed create refresh token", "error", err)
		return response.Tokens{}, errors.New("failed create refresh token")
	}

	session.RefreshToken = s.hasher.HashRefreshToken(res.RefreshToken)
	session.UUID = uid
	session.ExpiresAt = time.Now().Add(s.refreshTokenTTL)

	if isNewSession {
		err = s.repository.CreateSession(session)
		if err != nil {
			slog.Error("failed create session for user", "error", err)
			return response.Tokens{}, errors.New("failed create and save session")
		}
		slog.Debug("session created successfully", "user_id", session.UserID, "uuid", uid)
	} else {
		err = s.repository.UpdateSession(session)
		if err != nil {
			slog.Error("failed update session for user", "error", err)
			return response.Tokens{}, errors.New("failed update session")
		}

		slog.Debug("session updated successfully", "user_id", session.UserID, "uuid", uid)
	}
	return res, nil

}

// confirm code function need and after create session

func (s *AuthService) Logout(userId, uuid string) error {
	slog.Debug("logout from session", "uuid", uuid, "user_id", userId)

	err := s.repository.DeleteSessionByUuid(uuid)
	if err != nil {
		slog.Error("failed logout", "error", err.Error())
		return errors.New("failed logout")
	}
	return nil
}
