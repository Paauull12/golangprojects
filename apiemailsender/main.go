package main

import (
	"net/http"
	"encoding/base64"
	"net/smtp"
	"os"
	"time"
	"strings"
	"strconv"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Firma struct {
	Nume   string `json:"numeFirma"`
	CUICNP string `json:"cuicnp"`
	Email  string `json:"email"`
	Adresa string `json:"adresa"`
}

type Produs struct {
	Nume      string  `json:"numeProd"`
	Cantitate string  `json:"cantitate"`
	Pret      float32 `json:"pret"`
}

type InputUser struct {
	Sender   Firma     `json:"sender"`
	Receiver Firma     `json:"receiver"`
	Produse  []Produs  `json:"produse"`
	Seria    string    `json:"seria"`
	Culoare  string    `json:"culoare"`
	Data     time.Time `json:"data"`
}

func generateProductRows(products []Produs, color string) string {
	var rows []string
	for i, prod := range products {
		bgColor := "#ffffff"
		if i%2 == 1 {
			bgColor = "#f2f2f2"
		}
		
		row := fmt.Sprintf(`
			<tr style="background-color: %s; transition: background-color 0.2s;">
				<td style="padding: 10px 8px; border-bottom: 1px solid #ddd; text-align: left;">%s</td>
				<td style="padding: 10px 8px; border-bottom: 1px solid #ddd; text-align: center;">%s</td>
				<td style="padding: 10px 8px; border-bottom: 1px solid #ddd; text-align: right; font-weight: bold;">%.2f</td>
			</tr>`,
			bgColor,
			prod.Nume,
			prod.Cantitate,
			prod.Pret,
		)
		rows = append(rows, row)
	}
	
	if len(products) > 0 {
		var total float32
		for _, prod := range products {
			
			qty := 1.0
			if qtyFloat, err := strconv.ParseFloat(prod.Cantitate, 32); err == nil {
				qty = qtyFloat
			}
			total += prod.Pret * float32(qty)
		}
		
		totalRow := fmt.Sprintf(`
			<tr>
				<td colspan="2" style="padding: 12px 8px; text-align: right; font-weight: bold; border-top: 2px solid %s;">Total:</td>
				<td style="padding: 12px 8px; text-align: right; font-weight: bold; border-top: 2px solid %s;">%.2f</td>
			</tr>`,
			color, color, total,
		)
		rows = append(rows, totalRow)
	}
	
	return strings.Join(rows, "\n")
}

func sendEmail(to []string, input InputUser) error {

	auth := smtp.PlainAuth(
		"",
		os.Getenv("FROM_EMAIL"),
		os.Getenv("FROM_EMAIL_PASSWORD"),
		os.Getenv("FROM_EMAIL_SMTP"),
	)

	boundary := "invoice-boundary"
	subject := fmt.Sprintf("Subject: Invoice from %s\r\n", input.Sender.Nume)
	headers := "MIME-Version: 1.0\r\n"
	headers += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary)

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Invoice from %s</title>
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 800px; margin: 0 auto; padding: 20px;">
			<!-- Header with company color -->
			<div style="background-color: %s; color: white; padding: 20px; border-radius: 5px 5px 0 0; text-align: center;">
				<h1 style="margin: 0; font-size: 24px;">Invoice from %s</h1>
			</div>
			
			<!-- Main content with subtle border -->
			<div style="border: 1px solid #ddd; border-top: none; padding: 20px; border-radius: 0 0 5px 5px; background-color: #f9f9f9;">
				<div style="margin-bottom: 25px;">
					<p style="font-size: 16px; margin-bottom: 15px;">Dear <strong>%s</strong>,</p>
					<p style="font-size: 16px;">You have received an invoice from <strong>%s</strong>.</p>
				</div>
				
				<!-- Invoice Details Section -->
				<div style="background-color: white; padding: 15px; border-radius: 5px; margin-bottom: 20px; border-left: 4px solid %s;">
					<h2 style="margin-top: 0; color: %s; font-size: 18px;">Invoice Details</h2>
					<p style="margin: 8px 0;"><strong>Invoice Number:</strong> %s</p>
					<p style="margin: 8px 0;"><strong>Invoice Date:</strong> %s</p>
				</div>
				
				<!-- Sender Information Section -->
				<div style="background-color: white; padding: 15px; border-radius: 5px; margin-bottom: 20px; border-left: 4px solid %s;">
					<h2 style="margin-top: 0; color: %s; font-size: 18px;">Sender Information</h2>
					<p style="margin: 8px 0;"><strong>%s</strong></p>
					<p style="margin: 8px 0;">%s</p>
					<p style="margin: 8px 0;"><strong>Email:</strong> %s</p>
					<p style="margin: 8px 0;"><strong>CUICNP:</strong> %s</p>
				</div>
				
				<!-- Products Section -->
				<div style="background-color: white; padding: 15px; border-radius: 5px; margin-bottom: 20px;">
					<h2 style="margin-top: 0; color: %s; font-size: 18px;">Products</h2>
					<table style="width: 100%%; border-collapse: collapse; margin-top: 10px;">
						<thead>
							<tr>
								<th style="padding: 12px 8px; text-align: left; background-color: %s; color: white; border-radius: 5px 0 0 0;">Product</th>
								<th style="padding: 12px 8px; text-align: center; background-color: %s; color: white;">Quantity</th>
								<th style="padding: 12px 8px; text-align: right; background-color: %s; color: white; border-radius: 0 5px 0 0;">Price</th>
							</tr>
						</thead>
						<tbody>
							%s
						</tbody>
					</table>
				</div>
				
				<!-- Footer message -->
				<div style="margin-top: 30px; padding-top: 15px; border-top: 1px solid #ddd; color: #555;">
					<p style="margin: 5px 0;">Thank you for doing business with us. If you received this email by mistake, please reply and let us know.</p>
					<p style="margin: 20px 0 5px 0;">Best regards,</p>
					<p style="margin: 5px 0;"><strong>%s Team</strong></p>
				</div>
			</div>
			
			<!-- Small footer -->
			<div style="text-align: center; margin-top: 20px; font-size: 12px; color: #777;">
				<p>This is an automated invoice message. Please do not reply directly to this email.</p>
			</div>
		</body>
		</html>`, 
		input.Sender.Nume, 
		input.Culoare, 
		input.Sender.Nume, 
		input.Receiver.Nume, 
		input.Sender.Nume,
		input.Culoare,
		input.Culoare, 
		input.Seria,
		input.Data.Format("2006-01-02"),
		input.Culoare,
		input.Culoare,
		input.Sender.Nume, 
		input.Sender.Adresa, 
		input.Sender.Email, 
		input.Sender.CUICNP,
		input.Culoare,
		input.Culoare,
		input.Culoare,
		input.Culoare,
    	generateProductRows(input.Produse, input.Culoare),
    	input.Sender.Nume)

	pdfData, err := GeneratePDF(input)

	if err != nil {
		return err
	}

	message := subject + headers + "\r\n"
	
	message += fmt.Sprintf("--%s\r\n", boundary)
	message += "Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"
	message += htmlBody + "\r\n\r\n"
	
	message += fmt.Sprintf("--%s\r\n", boundary)
	message += "Content-Type: application/pdf\r\n"
	message += "Content-Disposition: attachment; filename=invoice.pdf\r\n"
	message += "Content-Transfer-Encoding: base64\r\n\r\n"
	message += base64.StdEncoding.EncodeToString(pdfData) + "\r\n"
	
	message += fmt.Sprintf("--%s--", boundary)
	
	return smtp.SendMail(
		os.Getenv("FROM_ADDR"),
		auth,
		os.Getenv("FROM_EMAIL"),
		to,
		[]byte(message),
	)
}

func main() {

	godotenv.Load()

	r := gin.Default()

	r.POST("/sendemail", func(c *gin.Context) {

		var input InputUser

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return 
		}

		err := sendEmail([]string{input.Receiver.Email}, input)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return 
		}


		c.JSON(http.StatusOK, gin.H{"msg": "ok bro :)"})
	})

	r.Run(":8888")
}
