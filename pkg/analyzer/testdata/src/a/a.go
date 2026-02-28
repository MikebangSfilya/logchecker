package a

import (
	"log/slog"

	"go.uber.org/zap"
)

func main() {
	slog.Info("Starting server") // want "the log message must begin with a lowercase letter"
	slog.Info("starting server") // OK
	slog.Info("123 server")      // OK (цифры в начале разрешены)
	slog.Info("")                // OK (пустая строка)

	slog.Error("ошибка")      // want "the log message must be in English only"
	slog.Info("error ошибка") // want "the log message must be in English only"
	slog.Info("error 123")    // OK

	slog.Warn("failed!!!")               // want "the log message must not contain special characters or emojis"
	slog.Warn("failed?")                 // want "the log message must not contain special characters or emojis"
	slog.Warn("wait...")                 // want "the log message must not contain special characters or emojis"
	slog.Warn("failed.service.")         // want "the log message must not contain special characters or emojis"
	slog.Warn("failed.service.........") // want "the log message must not contain special characters or emojis"
	slog.Warn("fire 🔥")                  // want "the log message must not contain special characters or emojis"
	slog.Warn("math: a + b")             // want "the log message must not contain special characters or emojis"

	slog.Info("server on 127.0.0.1") // OK (внутренние точки)
	slog.Info("version 1.2.3-beta")  // OK (точки и дефисы)
	slog.Info("path/to/file.go")     // OK (слеши и расширения)
	slog.Info("failed.service")      // OK (одиночная точка внутри)
	slog.Info("user_id: 123")        // OK (если разрешено двоеточие и подчеркивание)

	slog.Info("using api_key") // want "log message contains sensitive data: api_key"
	slog.Info("Token is set")  // want "the log message must begin with a lowercase letter" "log message contains sensitive data: token"
	slog.Info("token is set")  // want "log message contains sensitive data: token"

	password := "qwerty"
	secretKey := "12345"
	userToken := "abc"

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Error("bad password: " + password)    // want "log message contains sensitive data: password" "attempt to log sensitive variable: password"
	logger.Debug("failed to parse " + secretKey) // want "attempt to log sensitive variable: secret"
	logger.Info("send " + userToken)             // want "attempt to log sensitive variable: token"

	logger.Error("Database connection failed") // want "the log message must begin with a lowercase letter"
	logger.Info("запуск")                      // want "the log message must be in English only"
	logger.Info("request processed")           // OK
}
