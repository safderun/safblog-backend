package services

import (
	"errors"
	"fmt"
	"os"
	"safblog-backend/config"
	"safblog-backend/database"
	"safblog-backend/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(registeredUser models.RegisterModel) (models.Response, error) {
	fmt.Printf("Register running username: %s, email: %s, password: %s\n", registeredUser.Username, registeredUser.Email, registeredUser.Password)

	user := models.User{
		Username: registeredUser.Username,
		Email:    registeredUser.Email,
		Password: registeredUser.Password,
	}

	db := database.DB.Db

	var dbEmailUser models.User
	db.Find(&dbEmailUser, "email = ?", registeredUser.Email)
	if dbEmailUser.Email != "" {
		err := "email already in use"
		return models.Response{Message: "failed to create user", Error: err}, errors.New(err)
	}

	db.Find(&dbEmailUser, "id = ?", registeredUser.Username)
	if dbEmailUser.Email != "" {
		err := "username already in use"
		return models.Response{Message: "failed to create user", Error: err}, errors.New(err)
	}

	hash, err := saltAndHash(registeredUser.Password)
	if err != nil {
		error := "error while hashing the password"
		return models.Response{Message: "failed to has password", Error: error}, errors.New(error)
	}
	user.Password = hash

	err = db.Create(&user).Error
	if err != nil {
		error := "could not create user"
		return models.Response{Message: "failed to create user", Error: error}, errors.New(error)
	}

	return models.Response{Message: "user created", Data: fiber.Map{"message": "user created."}}, nil
}

func saltAndHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hash), err
}

func verifyPassword(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func LoginUser(loginUser models.LoginUser) (models.Response, error) {
	fmt.Printf("%s is logging in.\n", loginUser.Email)
	db := database.DB.Db

	var dbUser models.User

	db.Find(&dbUser, "email = ?", loginUser.Email)

	if dbUser.ID == uuid.Nil {
		err := "user not found"
		return CreateResponse("failed to find user", nil, err), errors.New(err)
		//return models.Response{Message: "failed to find user", Error: err}, errors.New(err)
	}

	isPasswordValid := verifyPassword(dbUser.Password, []byte(loginUser.Password))

	if !isPasswordValid {
		err := "credentials are not valid"
		return models.Response{Message: "failed to authenticate user", Error: err}, errors.New(err)
	}
	jwtHour, err := strconv.Atoi(os.Getenv("JWTHOUR"))
	if err != nil {
		err := "JWTHOUR env value is not integer"
		fmt.Println(err)
		return CreateResponse("internal server error", nil, ""), errors.New(err)
	}
	claims := jwt.MapClaims{
		"id":       dbUser.ID,
		"email":    dbUser.Email,
		"username": dbUser.Username,
		"isAdmin":  dbUser.IsAdmin,
		"exp":      time.Now().Add(time.Hour * time.Duration(jwtHour)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Generate encoded token and send it as response.

	t, err := token.SignedString([]byte(os.Getenv("JWTSECRET")))
	if err != nil {
		return models.Response{Message: "failed to create jwt signed token", Error: err.Error()}, errors.New(err.Error())
	}

	return models.Response{
		Message: "user login success",
		Data: fiber.Map{
			"token": t,
		},
	}, nil
}

func CreateRootUser() (models.Response, error) {
	fmt.Println("Creating root user.")
	rootUserUsername := config.Config("ROOT_USER_USERNAME")
	rootUserEmail := config.Config("ROOT_USER_EMAIL")
	rootUserPassword := config.Config("ROOT_USER_PASSWORD")

	user := models.User{
		Username: rootUserUsername,
		Email:    rootUserEmail,
		Password: rootUserPassword,
		IsAdmin:  true,
	}

	db := database.DB.Db

	var dbEmailUser models.User
	db.Find(&dbEmailUser, "email = ?", rootUserEmail)
	if dbEmailUser.Email != "" {
		err := "root user email already exists"
		fmt.Println("")
		return models.Response{Message: "failed to create root user", Error: err}, nil
	}

	hash, err := saltAndHash(rootUserPassword)
	if err != nil {
		error := "error while hashing the root user password"
		return models.Response{Message: "failed to hash root user password", Error: error}, errors.New(error)
	}
	user.Password = hash

	err = db.Create(&user).Error
	if err != nil {
		error := "could not create the root user"
		return models.Response{Message: "failed to create root user", Error: error}, errors.New(error)
	}

	return SuccessResponse("admin user created", fiber.Map{"message": "admin user created"}), nil
	//return services.SuccessResponse("admin user created")
}
