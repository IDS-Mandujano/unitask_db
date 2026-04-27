package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"unitask-api/internal/handler"
	"unitask-api/internal/repository"
	"unitask-api/internal/service"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Middleware para CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Cargar .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error cargando el archivo .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := repository.NewMySQLUserRepository(db)
	subjectRepo := repository.NewmysqlSubject(db)
	taskRepo := repository.NewmysqlTask(db)
	notificationRepo := repository.NewMySQLNotificationRepository(db)
	attachmentRepo := repository.NewMySQLAttachmentRepository(db)

	firebaseCredentialsPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if firebaseCredentialsPath == "" {
		firebaseCredentialsPath = "services-account.json"
	}

	notificationSender, err := service.NewFirebaseSender(firebaseCredentialsPath)
	if err != nil {
		log.Fatalf("Error inicializando Firebase sender: %v", err)
	}

	// 2. Handlers [cite: 112]
	authHandler := &handler.AuthHandler{Repo: userRepo}
	subjectHandler := &handler.SubjectHandler{Repo: subjectRepo}
	taskHandler := &handler.TaskHandler{Repo: taskRepo}
	notificationHandler := &handler.NotificationHandler{Repo: notificationRepo, Sender: notificationSender}
	attachmentHandler := &handler.AttachmentHandler{Repo: attachmentRepo}
	reminderTimezone := os.Getenv("REMINDER_TIMEZONE")
	if reminderTimezone == "" {
		reminderTimezone = "America/Mexico_City"
	}
	schedulerLoc, err := time.LoadLocation(reminderTimezone)
	if err != nil {
		log.Printf("timezone inválida '%s', usando Local: %v", reminderTimezone, err)
		schedulerLoc = time.Local
	}

	reminderWindowHours := 60
	if raw := os.Getenv("REMINDER_WINDOW_HOURS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			reminderWindowHours = parsed
		} else {
			log.Printf("REMINDER_WINDOW_HOURS inválido '%s', usando %d", raw, reminderWindowHours)
		}
	}

	schedulerInterval := time.Minute
	if raw := os.Getenv("REMINDER_INTERVAL_SECONDS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			schedulerInterval = time.Duration(parsed) * time.Second
		} else {
			log.Printf("REMINDER_INTERVAL_SECONDS inválido '%s', usando %s", raw, schedulerInterval)
		}
	}

	reminderScheduler := service.NewReminderScheduler(notificationRepo, notificationSender, reminderWindowHours, schedulerLoc)
	reminderScheduler.Start(schedulerInterval)

	mux := http.NewServeMux()

	// Auth - MVP 01 [cite: 84]
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.Handle("/auth/profile", handler.AuthMiddleware(http.HandlerFunc(authHandler.Profile)))

	// Subjects - MVP 02 [cite: 89]
	mux.Handle("/subjects", handler.AuthMiddleware(http.HandlerFunc(subjectHandler.HandleSubjects)))
	mux.Handle("/subjects/", handler.AuthMiddleware(http.HandlerFunc(subjectHandler.HandleSubjectByID)))

	// Tasks - MVP 03 [cite: 94]
	// En tu main.go, cambia la línea de /tasks por esto:
	// --- SECCIÓN DE TAREAS (MVP 03) ---

	mux.Handle("GET /tasks", handler.AuthMiddleware(http.HandlerFunc(taskHandler.GetTasks)))
	mux.Handle("POST /tasks", handler.AuthMiddleware(http.HandlerFunc(taskHandler.CreateTask)))
	mux.Handle("PATCH /tasks/complete", handler.AuthMiddleware(http.HandlerFunc(taskHandler.CompleteTask)))
	mux.Handle("PATCH /tasks/pending", handler.AuthMiddleware(http.HandlerFunc(taskHandler.MarkPendingTask)))
	mux.Handle("PUT /tasks", handler.AuthMiddleware(http.HandlerFunc(taskHandler.UpdateTask)))
	mux.Handle("DELETE /tasks", handler.AuthMiddleware(http.HandlerFunc(taskHandler.DeleteTask)))
	mux.Handle("/tasks/complete", handler.AuthMiddleware(http.HandlerFunc(taskHandler.CompleteTask)))
	mux.Handle("/tasks/pending", handler.AuthMiddleware(http.HandlerFunc(taskHandler.MarkPendingTask)))

	// Notifications - Firebase Push
	mux.Handle("POST /notifications/device-token", handler.AuthMiddleware(http.HandlerFunc(notificationHandler.RegisterDeviceToken)))
	mux.Handle("POST /notifications/device-token/remove", handler.AuthMiddleware(http.HandlerFunc(notificationHandler.UnregisterDeviceToken)))
	mux.Handle("POST /notifications/reminders/pending", handler.AuthMiddleware(http.HandlerFunc(notificationHandler.SendPendingReminder)))
	mux.Handle("POST /notifications/reminders/due-soon", handler.AuthMiddleware(http.HandlerFunc(notificationHandler.SendDueSoonReminder)))
	mux.Handle("POST /notifications/test", handler.AuthMiddleware(http.HandlerFunc(notificationHandler.SendTestNotification)))

	// Attachments - Firebase Storage metadata
	mux.Handle("GET /attachments", handler.AuthMiddleware(http.HandlerFunc(attachmentHandler.ListAttachments)))
	mux.Handle("POST /attachments", handler.AuthMiddleware(http.HandlerFunc(attachmentHandler.RegisterAttachment)))
	mux.Handle("DELETE /attachments", handler.AuthMiddleware(http.HandlerFunc(attachmentHandler.DeleteAttachment)))

	// ... arriba están tus handlers ...

	fmt.Println("-------------------------------------------")
	fmt.Println("✅ Servidor UniTask iniciado en el puerto " + port)
	fmt.Println("🚀 Conectado a la base de datos unitask_db")
	fmt.Println("🛠️  Esperando peticiones de la App...")
	fmt.Println("-------------------------------------------")

	log.Fatal(http.ListenAndServe(":"+port, enableCORS(mux)))
}
