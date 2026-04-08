package handlers

import (
	"errors"
	"fmt"
	"io"
	"loan-app/config"
	"loan-app/models"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gofiber/fiber/v2"
)

// Request struct for search
type SearchCarRequest struct {
	CarCode string `json:"car_code"`
}

func SearchCar(c *fiber.Ctx) error {
	var req SearchCarRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.CarCode == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Car Code is required",
		})
	}

	var car models.RedbookII
	result := config.DB.Where("carCode = ?", req.CarCode).First(&car)

	if result.Error != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Car not found",
		})
	}

	return c.JSON(car)
}

// Struct for Insurance Calculation Request
type CalculateInsuranceReq struct {
	LoanID           uint   `json:"loan_id"`
	InsuranceCompany string `json:"insurance_company"`
	Installments     uint   `json:"installments"`
	Age              int    `json:"age"`                // Optional: If provided, usage overrides calculation
	ContractSignDate string `json:"contract_sign_date"` // Contract sign date from the form
}

func CalculateInsuranceRate(c *fiber.Ctx) error {
	var req CalculateInsuranceReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	loan, err := requireLoanAccess(c, req.LoanID)
	if err != nil {
		if err == fiber.ErrForbidden {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}
		if err == fiber.ErrNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Loan not found"})
		}
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	// 1. Gender: Male=1, Female=2
	genderCode := 0
	if loan.Gender == "male" {
		genderCode = 1
	} else {
		genderCode = 0
	}

	// 2. Age Calculation
	var age int
	var signDate time.Time
	var parseErr error

	// Priority: req.ContractSignDate (form) -> loan.ContractSignDate (DB) -> time.Now()
	if req.ContractSignDate != "" {
		// Accept multiple date formats sent by the frontend.
		formats := []string{"2006-01-02", "02-01-2006", "02/01/2006"}
		parsed := false
		for _, f := range formats {
			signDate, parseErr = time.Parse(f, req.ContractSignDate)
			if parseErr == nil {
				parsed = true
				break
			}
		}
		if !parsed {
			signDate = time.Now()
		}
	} else if loan.ContractSignDate != nil && *loan.ContractSignDate != "" {
		signDate, parseErr = time.Parse("2006-01-02", *loan.ContractSignDate)
		if parseErr != nil {
			signDate = time.Now()
		}
	} else {
		signDate = time.Now()
	}

	if req.Age > 0 {
		age = req.Age
	} else {
		// Calculate Server Side if Age not provided
		// DateOfBirth is now *string
		dobStr := ""
		if loan.DateOfBirth != nil {
			dobStr = *loan.DateOfBirth
		}

		if dobStr != "" {
			dob, parseErr := time.Parse("2006-01-02", dobStr)
			if parseErr == nil {
				age = int(signDate.Year() - dob.Year())
				if signDate.YearDay() < dob.YearDay() {
					age--
				}
				if age < 0 {
					age = 0
				}
			} else {
				age = 0
			}
		} else {
			age = 0
		}
	}

	// 3. Call DB Function
	var rate float64

	// Format Date to YYYYMMDD
	signDateStr := signDate.Format("20060102")

	// Convert numbers to strings as per example: Fnc_LoanProtect_Rate('03','1','48','41','20250226')
	genderStr := fmt.Sprintf("%d", genderCode)
	installmentsStr := fmt.Sprintf("%d", req.Installments)
	ageStr := fmt.Sprintf("%d", age)

	// SQL: SELECT Fnc_LoanProtect_Rate(?, ?, ?, ?, ?) AS RATE
	query := "SELECT Fnc_LoanProtect_Rate(?, ?, ?, ?, ?) AS RATE"

	// Use Row() for scalar scan to avoid GORM slice issues
	err = config.DB.Raw(query,
		req.InsuranceCompany,
		genderStr,
		installmentsStr,
		ageStr,
		signDateStr,
	).Row().Scan(&rate)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Calculation failed: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"rate": rate,
	})
}

// Request struct for Agent Search
type SearchAgentRequest struct {
	Query string `json:"query"`
}

func SearchAgent(c *fiber.Ctx) error {
	var req SearchAgentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Query == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Query is required",
		})
	}

	var agents []models.LoanProtectLicense
	// Search by EmpId or EmpName using LIKE
	// Note: Adjust wildcards as per DB requirement (e.g. %query%)
	searchTerm := "%" + req.Query + "%"
	result := config.DB.Where("EmpId LIKE ? OR EmpName LIKE ?", searchTerm, searchTerm).Limit(20).Find(&agents)

	if result.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error: " + result.Error.Error(),
		})
	}

	return c.JSON(agents)
}

// Request struct for Showroom Search
type SearchShowroomRequest struct {
	Query string `json:"query"`
}

func SearchShowroom(c *fiber.Ctx) error {
	var req SearchShowroomRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Query == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Query is required",
		})
	}

	var showrooms []models.Showroom
	searchTerm := "%" + req.Query + "%"
	// Search by ID or Name
	result := config.DB.Where("ShowRoomId LIKE ? OR ShowRoomName LIKE ?", searchTerm, searchTerm).Limit(20).Find(&showrooms)

	if result.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error: " + result.Error.Error(),
		})
	}

	return c.JSON(showrooms)
}

// GetTitles fetches all titles from the database
func GetTitles(c *fiber.Ctx) error {
	var titles []models.Title
	if err := config.DB.Find(&titles).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch titles",
		})
	}
	return c.JSON(titles)
}

// GetRaces fetches all races from the database
func GetRaces(c *fiber.Ctx) error {
	var races []models.Race
	if err := config.DB.Find(&races).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch races",
		})
	}
	return c.JSON(races)
}

// GetNations fetches all nations from the database
func GetNations(c *fiber.Ctx) error {
	var nations []models.Nation
	if err := config.DB.Find(&nations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch nations",
		})
	}
	return c.JSON(nations)
}

// GetReligions fetches all religions from the database
func GetReligions(c *fiber.Ctx) error {
	var religions []models.Religion
	if err := config.DB.Find(&religions).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch religions",
		})
	}
	return c.JSON(religions)
}

// GetSituations fetches all marital statuses from the database
func GetSituations(c *fiber.Ctx) error {
	var situations []models.Situation
	if err := config.DB.Find(&situations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch situations",
		})
	}
	return c.JSON(situations)
}

// GetOccupations fetches all occupations from the database
func GetOccupations(c *fiber.Ctx) error {
	var occupations []models.Occupy
	if err := config.DB.Find(&occupations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch occupations",
		})
	}
	return c.JSON(occupations)
}

// GetInsuComps fetches all insurance companies from the database
func GetInsuComps(c *fiber.Ctx) error {
	var companies []models.InsuComp
	if err := config.DB.Find(&companies).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch insurance companies",
		})
	}
	return c.JSON(companies)
}

// GetInsuClasses fetches all insurance classes from the database
func GetInsuClasses(c *fiber.Ctx) error {
	var classes []models.InsuClass
	if err := config.DB.Find(&classes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch insurance classes",
		})
	}
	return c.JSON(classes)
}

var (
	r2Once                      sync.Once
	r2Client                    *s3.S3
	allowedInsuranceUploadTypes = map[string]string{
		".pdf":  "application/pdf",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
	}
	presignFileURL = func(filename string) (string, error) {
		svc := getR2Client()
		req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(config.GetConfig().R2BucketName),
			Key:    aws.String(filename),
		})

		return req.Presign(15 * time.Minute)
	}
	putR2Object = func(filename string, body io.ReadSeeker, contentType string) error {
		svc := getR2Client()
		_, err := svc.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String(config.GetConfig().R2BucketName),
			Key:         aws.String(filename),
			Body:        body,
			ContentType: aws.String(contentType),
		})
		return err
	}
)

// getR2Client returns a cached S3 client (singleton) for Cloudflare R2.
func getR2Client() *s3.S3 {
	r2Once.Do(func() {
		r2 := config.GetConfig()
		creds := credentials.NewStaticCredentials(r2.R2AccessKeyId, r2.R2SecretAccessKey, "")
		awsCfg := aws.NewConfig().
			WithRegion("auto").
			WithEndpoint(r2.R2Endpoint).
			WithCredentials(creds)
		sess := session.Must(session.NewSession(awsCfg))
		r2Client = s3.New(sess)
	})
	return r2Client
}

func validateInsuranceUpload(fileHeader *multipart.FileHeader, maxBytes int64) (multipartFile io.ReadSeeker, contentType string, err error) {
	if fileHeader == nil {
		return nil, "", errors.New("file is required")
	}
	if fileHeader.Size <= 0 {
		return nil, "", errors.New("file is empty")
	}
	if fileHeader.Size > maxBytes {
		return nil, "", fmt.Errorf("file exceeds maximum size of %d bytes", maxBytes)
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	expectedType, ok := allowedInsuranceUploadTypes[ext]
	if !ok {
		return nil, "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, "", errors.New("failed to open file")
	}

	sniffer, ok := src.(io.ReadSeeker)
	if !ok {
		_ = src.Close()
		return nil, "", errors.New("uploaded file does not support seeking")
	}

	header := make([]byte, 512)
	n, readErr := sniffer.Read(header)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		_ = src.Close()
		return nil, "", errors.New("failed to inspect file content")
	}

	detectedType := http.DetectContentType(header[:n])
	if detectedType != expectedType {
		_ = src.Close()
		return nil, "", fmt.Errorf("detected content type %s does not match %s", detectedType, expectedType)
	}

	if _, err := sniffer.Seek(0, io.SeekStart); err != nil {
		_ = src.Close()
		return nil, "", errors.New("failed to rewind uploaded file")
	}

	return sniffer, detectedType, nil
}

// Upload Insurance File to Cloudflare R2
func UploadInsuranceFile(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		if err == fiber.ErrForbidden {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}
		if err == fiber.ErrNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Loan not found"})
		}
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "File upload failed"})
	}

	src, contentType, err := validateInsuranceUpload(fileHeader, config.GetConfig().UploadMaxFileSizeBytes)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if closer, ok := src.(io.Closer); ok {
		defer closer.Close()
	}

	// Generate safe key (filename)
	filename := fmt.Sprintf("%s_%d_%s", loanID, time.Now().UnixNano(), fileHeader.Filename)

	// Upload to R2
	if err := putR2Object(filename, src, contentType); err != nil {
		log.Printf("R2 Upload Error: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upload to Cloud Storage"})
	}

	// Update Database

	if loan.CarInsuranceFile == "" {
		loan.CarInsuranceFile = filename
	} else {
		loan.CarInsuranceFile = loan.CarInsuranceFile + "," + filename
	}

	config.DB.Save(&loan)
	if err := createLoanFileMetadata(&models.LoanFile{
		LoanID:       loan.ID,
		StorageKey:   filename,
		OriginalName: fileHeader.Filename,
		Category:     models.LoanFileCategoryCarInsurance,
		UploadedBy:   parseJWTUsername(c.Cookies("token")),
	}); err != nil {
		log.Printf("loan file metadata create error for %s: %v", filename, err)
	}

	return c.JSON(fiber.Map{
		"message":  "Upload success",
		"filename": filename,
		// Return API URL for viewing (via presigned redirect)
		"url": "/file/" + filename,
	})
}

// GetFile redirects to a presigned URL after verifying the current user can access the file.
func GetFile(c *fiber.Ctx) error {
	filename := c.Params("filename")
	if filename == "" {
		return c.Status(400).SendString("Filename required")
	}

	if _, err := requireFileAccess(c, filename); err != nil {
		if err == fiber.ErrForbidden {
			return c.Status(fiber.StatusForbidden).SendString("Forbidden")
		}
		if err == fiber.ErrUnauthorized {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}
		return c.Status(fiber.StatusNotFound).SendString("File not found")
	}

	urlStr, err := presignFileURL(filename)
	if err != nil {
		log.Printf("Presign Error: %v", err)
		return c.Status(500).SendString("Failed to generate download link")
	}

	return c.Redirect(urlStr)
}

// Delete Insurance File
type DeleteFileRequest struct {
	Filename string `json:"filename"`
}

func DeleteInsuranceFile(c *fiber.Ctx) error {
	loanID := c.Cookies("loan_id")
	if loanID == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	loan, err := requireLoanAccess(c, loanID)
	if err != nil {
		if err == fiber.ErrForbidden {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}
		if err == fiber.ErrNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Loan not found"})
		}
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req DeleteFileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Remove filename from DB field (comma separated)
	files := strings.Split(loan.CarInsuranceFile, ",")
	newFiles := []string{}
	found := false
	for _, f := range files {
		if f == req.Filename {
			found = true

			// Delete from R2
			svc := getR2Client()
			_, err := svc.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(config.GetConfig().R2BucketName),
				Key:    aws.String(f),
			})
			if err != nil {
				log.Printf("R2 Delete Error for %s: %v", f, err)
				// Continue to remove from DB anyway? Yes.
			}

		} else {
			if f != "" {
				newFiles = append(newFiles, f)
			}
		}
	}

	if found {
		loan.CarInsuranceFile = strings.Join(newFiles, ",")
		config.DB.Save(&loan)
		if err := deleteLoanFileMetadata(loan.ID, req.Filename); err != nil {
			log.Printf("loan file metadata delete error for %s: %v", req.Filename, err)
		}
		return c.JSON(fiber.Map{"message": "File deleted"})
	}

	return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "File not found in record"})
}

// DeleteLoan deletes a loan application and its guarantors
func DeleteLoan(c *fiber.Ctx) error {
	type Request struct {
		ID int `json:"id"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	loan, err := requireLoanAccess(c, req.ID)
	if err != nil {
		if err == fiber.ErrForbidden {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}
		if err == fiber.ErrNotFound {
			return c.Status(404).JSON(fiber.Map{"error": "Loan not found"})
		}
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	// 2. Delete Uploaded Files
	if loan.CarInsuranceFile != "" {
		files := strings.Split(loan.CarInsuranceFile, ",")
		svc := getR2Client()
		for _, f := range files {
			if f != "" {
				// Delete from R2
				_, err := svc.DeleteObject(&s3.DeleteObjectInput{
					Bucket: aws.String(config.GetConfig().R2BucketName),
					Key:    aws.String(f),
				})
				if err != nil {
					log.Printf("Failed to delete file %s from R2: %v", f, err)
				}
			}
		}
	}

	// Transaction to ensure atomicity
	tx := config.DB.Begin()

	// 3. Delete Guarantors associated with this LoanID
	if err := tx.Where("loan_id = ?", req.ID).Unscoped().Delete(&models.Guarantor{}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete guarantors"})
	}

	// 4. Delete Loan Application
	if err := tx.Delete(&models.LoanApplication{}, req.ID).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete loan application"})
	}

	tx.Commit()
	WriteAudit(c, "delete_loan", loan.RefCode, loan.FirstName+" "+loan.LastName)
	return c.JSON(fiber.Map{"success": true})
}
