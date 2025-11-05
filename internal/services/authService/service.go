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
	"github.com/sibhellyx/Messenger/pkg/auth"
)

type RepositoryInterface interface {
	CreateSession(session entity.Session) error
	CreateUser(user entity.User) error
	CreateUserAndGetId(user entity.User) (uint, error)
	DeleteSessionByUuid(uuid string) error
	FindJwtSessionByUuidAndRefreshToken(uuid string, refreshToken string) (*entity.Session, error)
	GetUserByCredentails(tgname string, password string) (*entity.User, error)
	UpdateSession(session entity.Session) error
	DeleteExpiredSessions(userId uint) error
	DeleteOldestSession(userId uint) error
	GetUserByTgname(tgname string) (*entity.User, error)
	CountActiveSessions(userId uint) (int64, error)
	CreateProfile(profile entity.UserProfile) error
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
	GetLinkForFinishRegister(tgName string) (string, string)
	SendCode(code string, userId uint) error
}

type RedisRepositoryInterface interface {
	SaveRegistrationToken(token string, user entity.User, ttl time.Duration) error
	GetRegistrationToken(token string) (entity.User, error)
	DeleteRegistrationToken(token string) error
	SaveLoginCode(userID uint, code string, ttl time.Duration) error
	GetLoginCode(userID uint) (string, error)
	IncrementLoginAttempts(userID uint) error
	SaveUserRegistration(userID uint, tgChatId int64) error
	GetUserRegistration(userID uint) (int64, error)
}

type AuthService struct {
	repository RepositoryInterface

	hasher       HasherInterface
	tokenManager TokenManagerInterface
	bot          BotServiceInterface
	redis        RedisRepositoryInterface

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	activeSessions  int
}

func NewAuthService(
	repository RepositoryInterface,
	hasher HasherInterface,
	manager TokenManagerInterface,
	redis RedisRepositoryInterface,
	accessTokenTTL, refreshTokenTTL time.Duration,
	activeSessions int,
) *AuthService {
	return &AuthService{
		repository:      repository,
		hasher:          hasher,
		tokenManager:    manager,
		redis:           redis,
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

	u, err := s.repository.GetUserByTgname(r.Tgname)
	if err == nil || u != nil {
		slog.Error("this user is already registered")
		return "", errors.New("this user is already registered")
	}
	// create user
	user := entity.User{
		Name:     r.Name,
		Surname:  r.Surname,
		Tgname:   r.Tgname,
		Password: r.Password,
	}

	// send to bot this user name and get link for register
	token, link := s.bot.GetLinkForFinishRegister(r.Tgname)
	// save to reddis user
	err = s.redis.SaveRegistrationToken(token, user, 5*time.Minute)
	if err != nil {
		slog.Error("failed save user in repo for checking tokens of registration", "error", err)
		return "", errors.New("failed save user")
	}

	slog.Debug("service register completed")
	return link, nil
}

func (s *AuthService) GetTokenFromRedis(token string) (entity.User, error) {
	return s.redis.GetRegistrationToken(token)
}

func (s *AuthService) DeleteRegistrationTokenFromRedis(token string) error {
	return s.redis.DeleteRegistrationToken(token)
}

func (s *AuthService) SaveUserRegistration(userId uint, tgChatId int64) error {
	return s.redis.SaveUserRegistration(userId, tgChatId)
}

func (s *AuthService) GetUserRegistration(userID uint) (int64, error) {
	return s.redis.GetUserRegistration(userID)
}

// activate user account after start tg bot
func (s *AuthService) Activate(user entity.User) (uint, error) {
	slog.Debug("activating user", "tgName", user.Tgname)
	id, err := s.repository.CreateUserAndGetId(user)
	if err != nil {
		slog.Error("failed create user", "tgName", user.Tgname)
		return 0, err
	}
	profile := entity.UserProfile{
		UserID:      id,
		Bio:         "",
		Avatar:      "",
		DateOfBirth: nil,
	}
	err = s.repository.CreateProfile(profile)
	if err != nil {
		slog.Error("failed create profile for user", "tgName", user.Tgname)
	}
	slog.Debug("activating user completed", "tgName", user.Tgname)
	return id, nil
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
		return 0, errors.New("invalid credentials")
	}

	// generate code for verify login
	code := auth.GenerateLoginCode()
	// sending code in telegram
	err = s.bot.SendCode(code, u.ID)
	if err != nil {
		slog.Error("failed send code to user", "id", u.ID, "tgName", u.Tgname)
		return 0, errors.New("failed send code, try again later")
	}
	// save code in storage for check
	s.redis.SaveLoginCode(u.ID, code, 2*time.Minute)
	return u.ID, nil
}

func (s *AuthService) VerifyCode(req request.VerifyCodeRequest, params request.LoginParams) (response.Tokens, error) {
	slog.Debug("service of veryficate code started")
	err := req.Validate()
	if err != nil {
		slog.Error("error validating user input", "error", err.Error())
		return response.Tokens{}, err
	}
	id, err := strconv.ParseUint(req.UserID, 10, 32)
	if err != nil {
		slog.Error("failed parse user_id to uint", "user_id", req.UserID)
		return response.Tokens{}, errors.New("user_id incorrect")
	}

	// check exist this user in logining
	codeFromStorage, err := s.redis.GetLoginCode(uint(id))
	if err != nil {
		slog.Error("can't find code in storage for this user", "error", err)
		return response.Tokens{}, errors.New("code not found")
	}
	if codeFromStorage != req.Code {
		slog.Error("user not exist in storage of logining and code or code incorrect")
		return response.Tokens{}, errors.New("code incorrect")
	}
	err = s.redis.IncrementLoginAttempts(uint(id))
	if err != nil {
		slog.Warn("error in increment logining attemps", "error", err)
	}

	slog.Debug("service of veryficate code completed")
	return s.createSession(uint(id), params)
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
