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
}

type HasherInterface interface {
	Hash(password string) (string, error)
	HashRefreshToken(refreshToken string) string
}

type TokenManagerInterface interface {
	NewJWT(payload payload.JwtPayload, ttl time.Duration) (string, error)
	NewRefreshToken() (string, error)
	Parse(accessToken string) (payload.JwtPayload, error)
}

type AuthService struct {
	repository RepositoryInterface

	logger *slog.Logger

	hasher       HasherInterface
	tokenManager TokenManagerInterface

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(repository RepositoryInterface, logger *slog.Logger, hasher HasherInterface, manager TokenManagerInterface, accessTokenTTL, refreshTokenTTL time.Duration) *AuthService {
	return &AuthService{
		repository:      repository,
		logger:          logger,
		hasher:          hasher,
		tokenManager:    manager,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// register user service layer
func (s *AuthService) RegisterUser(user entity.User) error {
	s.logger.Debug("service register started")
	err := user.Validate() //function return error if user not isValid
	if err != nil {
		s.logger.Error("error validating user input", "error", err.Error())
		return err
	}
	user.Password, err = s.hasher.Hash(user.Password) //hash password
	if err != nil {
		s.logger.Error("failed hash password", "error", err.Error())
		return err
	}
	err = s.repository.CreateUser(user) //write to repo
	if err != nil {
		s.logger.Error("failed create user in repo", "error", err.Error())
		return err
	}
	s.logger.Debug("service register completed")
	return nil
}

func (s *AuthService) SignIn(user request.LoginRequest, params request.LoginParams) (response.Tokens, error) {
	s.logger.Debug("service login started")
	err := user.Validate()
	if err != nil {
		s.logger.Error("error validating user input", "error", err.Error())
		return response.Tokens{}, err
	}
	passwordHash, err := s.hasher.Hash(user.Password) //hash password
	if err != nil {
		s.logger.Error("failed hash password", "error", err.Error())
		return response.Tokens{}, errors.New("failed hashed password")
	}
	u, err := s.repository.GetUserByCredentails(user.Tgname, passwordHash)
	if err != nil {
		s.logger.Error("failed get user", "error", err.Error())
		return response.Tokens{}, errors.New("incorrect name or password")
	}
	return s.createSession(u.ID, params)
}

func (s *AuthService) RefreshToken(tokens response.Tokens, params request.LoginParams) (response.Tokens, error) {
	s.logger.Debug("service refresh token started")
	err := tokens.Validate()
	if err != nil {
		s.logger.Error("error validating tokens", "error", err.Error())
		return response.Tokens{}, err
	}
	refreshTokenHash := s.hasher.HashRefreshToken(tokens.RefreshToken)

	payload, err := s.tokenManager.Parse(tokens.AccessToken)
	if err != nil {
		s.logger.Error("failed parse access token", "error", err.Error())
		return response.Tokens{}, errors.New("failed parse access token")
	}

	session, err := s.repository.FindJwtSessionByUuidAndRefreshToken(payload.Uuid, refreshTokenHash)
	if err != nil {
		s.logger.Error("failed founding session with this token token", "error", err.Error())
		return response.Tokens{}, errors.New("failed found session")
	}
	if session == nil {
		s.logger.Debug("session with this uuid and refresh token not found", "user_id", payload.UserId, "uuid", payload.Uuid, "refresh_token", tokens.RefreshToken)
		return response.Tokens{}, errors.New("this session not found or this refresh token was issued separately")
	}
	if session.UserAgent != params.UserAgent {
		err := s.repository.DeleteSessionByUuid(payload.Uuid)
		if err != nil {
			s.logger.Error("failed deleted session, different user agents", "error", err.Error())
			return response.Tokens{}, errors.New("failed delete session, different user agents")
		}
		errMsg := "error refresh token from another user agent"
		s.logger.Error("different user agents", "error", errMsg)
		return response.Tokens{}, errors.New(errMsg)
	}
	return s.generateSessionAndSave(*session, false)
}

func (s *AuthService) createSession(userId uint, params request.LoginParams) (response.Tokens, error) {
	s.logger.Debug("creating session started")

	// add cheecking old sessions end expired, clear for free limit session
	if err := s.repository.DeleteExpiredSessions(userId); err != nil {
		s.logger.Warn("failed to cleanup expired sessions", "error", err)
	}

	activeCount, err := s.repository.CountActiveSessions(userId)
	if err != nil {
		return response.Tokens{}, fmt.Errorf("failed to count active sessions: %w", err)
	}

	const maxActiveSessions = 5 // add geting from config

	if activeCount >= maxActiveSessions {
		if err := s.repository.DeleteOldestSession(userId); err != nil {
			s.logger.Warn("failed to delete oldest session", "error", err)
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
	s.logger.Debug("generation session", "usesr_id", session.UserID)

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

	s.logger.Debug("creating access token")
	res.AccessToken, err = s.tokenManager.NewJWT(payload, s.accessTokenTTL)
	if err != nil {
		s.logger.Error("failed create access token", "error", err)
		return response.Tokens{}, errors.New("failed create access token")
	}

	s.logger.Debug("creating refresh token")
	res.RefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		s.logger.Error("failed create refresh token", "error", err)
		return response.Tokens{}, errors.New("failed create refresh token")
	}

	session.RefreshToken = s.hasher.HashRefreshToken(res.RefreshToken)
	session.UUID = uid
	session.ExpiresAt = time.Now().Add(s.refreshTokenTTL)

	if isNewSession {
		err = s.repository.CreateSession(session)
		if err != nil {
			s.logger.Error("failed create session for user", "error", err)
			return response.Tokens{}, errors.New("failed create and save session")
		}
		s.logger.Debug("session created successfully", "user_id", session.UserID, "uuid", uid)
	} else {
		err = s.repository.UpdateSession(session)
		if err != nil {
			s.logger.Error("failed update session for user", "error", err)
			return response.Tokens{}, errors.New("failed update session")
		}

		s.logger.Debug("session updated successfully", "user_id", session.UserID, "uuid", uid)
	}
	return res, nil

}

func (s *AuthService) Logout(userId, uuid string) error {
	s.logger.Debug("logout from session", "uuid", uuid, "user_id", userId)

	err := s.repository.DeleteSessionByUuid(uuid)
	if err != nil {
		s.logger.Error("failed logout", "error", err.Error())
		return errors.New("failed logout")
	}
	return nil
}
