package routes

import (
	"loan-app/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func Setup(app *fiber.App) {
	app.Use("/ws", handlers.WsHandler)
	app.Get("/ws", websocket.New(handlers.WsConnect))

	app.Get("/login", handlers.LoginPage)
	app.Post("/login", handlers.LoginPost)
	app.Get("/logout", handlers.Logout)
	app.Use(handlers.AuthMiddleware)

	app.Get("/change-password", handlers.ChangePasswordPage)
	app.Post("/change-password", handlers.ChangePasswordPost)

	app.Get("/", handlers.LoginPage)
	app.Get("/main", handlers.MainPage)
	app.Get("/step1", handlers.Step1)
	app.Post("/step1", handlers.Step1Post)
	app.Get("/step2", handlers.Step2)
	app.Post("/step2", handlers.Step2Post)
	app.Get("/step3", handlers.Step3)
	app.Post("/step3", handlers.Step3Post)
	app.Get("/step4", handlers.Step4)
	app.Post("/step4", handlers.Step4Post)
	app.Get("/step5", handlers.Step5)
	app.Post("/step5", handlers.Step5Post)
	app.Get("/step6", handlers.Step6)
	app.Post("/step6", handlers.Step6Post)
	app.Get("/step7", handlers.Step7)
	app.Post("/step7", handlers.Step7Post)

	// API
	app.Get("/api/loan-list", handlers.GetLoanList) // 📱 Mobile App
	app.Post("/api/sync-work", handlers.UpdateSyncStatus)
	app.Post("/api/update-status", handlers.UpdateStatus)
	app.Post("/api/search-car", handlers.SearchCar)
	app.Post("/api/calculate-insurance", handlers.CalculateInsuranceRate)
	app.Post("/api/search-agent", handlers.SearchAgent)
	app.Post("/api/search-showroom", handlers.SearchShowroom)
	app.Post("/api/upload-insurance-file", handlers.UploadInsuranceFile)
	app.Post("/api/delete-insurance-file", handlers.DeleteInsuranceFile)
	app.Post("/api/delete-loan", handlers.DeleteLoan)
	app.Get("/api/titles", handlers.GetTitles)
	app.Get("/api/races", handlers.GetRaces)
	app.Get("/api/nations", handlers.GetNations)
	app.Get("/api/religions", handlers.GetReligions)
	app.Get("/api/situations", handlers.GetSituations)
	app.Get("/api/occupations", handlers.GetOccupations)
	app.Get("/api/insucomps", handlers.GetInsuComps)
	app.Get("/api/insuclasses", handlers.GetInsuClasses)

	// OCR — Gemini AI วิเคราะห์เล่มทะเบียนรถไทย + บัตรประชาชน
	v1 := app.Group("/api/v1")
	v1.Post("/ocr/vehicle", handlers.OcrVehicleBook)
	v1.Post("/ocr/idcard", handlers.OcrIDCard)

	// File Download (R2)
	app.Get("/file/:filename", handlers.GetFile)

	// Guarantor
	app.Get("/add-guarantor", handlers.AddGuarantorGetV2)
	app.Post("/add-guarantor", handlers.AddGuarantorPostV2)
	app.Post("/delete-guarantor", handlers.DeleteGuarantor)
}
