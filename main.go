package main

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price        string `json:"price"`
}

type Receipt struct {
	StoreName      string `json:"retailer"`
	DateOfPurchase string `json:"purchaseDate"`
	TimeOfPurchase string `json:"purchaseTime"`
	Items          []Item `json:"items"`
	AmountTotal    string `json:"total"`
}

var Receipts = make(map[string]Receipt)

// POST handler
func ProcessReceipt(c *gin.Context) {
	var receipt Receipt
	if err := c.ShouldBindJSON(&receipt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	receiptID := uuid.New().String()
	Receipts[receiptID] = receipt
	c.JSON(http.StatusOK, gin.H{"id": receiptID})
}

func CalculatePoints(c *gin.Context) {
	receiptID := c.Param("id")
	receipt, exists := Receipts[receiptID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}
	points := 0
	for _, char := range receipt.StoreName {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			points++
		}
	}
	total, _ := strconv.ParseFloat(receipt.AmountTotal, 64)
	if math.Mod(total*100, 10.00) == 0 {
		points += 50
	}
	if math.Mod(total, 0.25) == 0 {
		points += 25
	}
	points += int(len(receipt.Items)/2) * 5
	for _, Item := range receipt.Items {
		if len(strings.Trim(Item.ShortDescription, " "))%3 == 0 {
			price, _ := strconv.ParseFloat(Item.Price, 64)
			points += int(math.Ceil(price * 0.2))
		}
	}
	date, _ := time.Parse("2006-01-02", receipt.DateOfPurchase)
	if date.Day()%2 != 0 {
		points += 6
	}
	purchaseTime, _ := time.Parse("15:04", receipt.TimeOfPurchase)
	if purchaseTime.Hour() >= 14 && purchaseTime.Hour() < 16 && purchaseTime.Minute() > 0 {
		points += 10
	}
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func main() {
	router := gin.Default()
	router.POST("/receipts/process", ProcessReceipt)
	router.GET("/receipt/:id/points", CalculatePoints)
	router.Run(":5000") // listen and serve on 0.0.0.0:5000
}
