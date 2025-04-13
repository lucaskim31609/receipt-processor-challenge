package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Receipt represents the JSON input structure.
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

// Item represents a single item on the receipt.
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// ValidatedReceiptData holds parsed data needed for point calculations.
type ValidatedReceiptData struct {
	Retailer      string
	PurchaseDate  time.Time
	PurchaseTime  time.Time
	Items         []ValidatedItemData
	Total         float64
	OriginalItems int
}

// ValidatedItemData holds parsed item data.
type ValidatedItemData struct {
	ShortDescription string
	Price            float64
}

// Validation regular expressions and helpers.
var (
	retailerRegex     = regexp.MustCompile(`^[\w\s\-&]+$`)
	priceTotalRegex   = regexp.MustCompile(`^\d+\.\d{2}$`)
	itemDescRegex     = regexp.MustCompile(`^[\w\s\-]+$`)
	idPatternRegex    = regexp.MustCompile(`^\S+$`)
	alphanumericCheck = func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }
)

// validateAndParseReceipt checks the input receipt's format and structure,
// returning parsed data or an error.
func validateAndParseReceipt(receipt *Receipt) (*ValidatedReceiptData, error) {
	if !retailerRegex.MatchString(receipt.Retailer) {
		return nil, fmt.Errorf("invalid retailer format")
	}
	purchaseDate, err := time.Parse("2006-01-02", receipt.PurchaseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid purchaseDate format (YYYY-MM-DD)")
	}
	purchaseTime, err := time.Parse("15:04", receipt.PurchaseTime)
	if err != nil {
		return nil, fmt.Errorf("invalid purchaseTime format (HH:MM)")
	}
	if !priceTotalRegex.MatchString(receipt.Total) {
		return nil, fmt.Errorf("invalid total format (N.NN)")
	}
	totalFloat, _ := strconv.ParseFloat(receipt.Total, 64)

	if receipt.Items == nil || len(receipt.Items) == 0 {
		return nil, fmt.Errorf("items array cannot be empty")
	}

	var validatedItems []ValidatedItemData
	for i, item := range receipt.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		if trimmedDesc == "" {
			return nil, fmt.Errorf("item %d: shortDescription required", i)
		}
		if !itemDescRegex.MatchString(item.ShortDescription) {
			return nil, fmt.Errorf("item %d: invalid shortDescription format", i)
		}

		if !priceTotalRegex.MatchString(item.Price) {
			return nil, fmt.Errorf("item %d: invalid price format (N.NN)", i)
		}
		priceFloat, _ := strconv.ParseFloat(item.Price, 64)
		validatedItems = append(validatedItems, ValidatedItemData{
			ShortDescription: item.ShortDescription,
			Price:            priceFloat,
		})
	}

	return &ValidatedReceiptData{
		Retailer:      receipt.Retailer,
		PurchaseDate:  purchaseDate,
		PurchaseTime:  purchaseTime,
		Items:         validatedItems,
		Total:         totalFloat,
		OriginalItems: len(receipt.Items),
	}, nil
}

// calculatePoints computes the points awarded based on the defined rules.
func calculatePoints(data *ValidatedReceiptData) int64 {
	var points int64 = 0
	const floatEpsilon = 0.0000001

	// Rule 1: Alphanumeric characters in retailer name
	retailerPoints := 0
	for _, r := range data.Retailer {
		if alphanumericCheck(r) {
			retailerPoints++
		}
	}
	points += int64(retailerPoints)

	// Rule 2: Round dollar total
	if math.Abs(data.Total-math.Trunc(data.Total)) < floatEpsilon && data.Total > 0 {
		points += 50
	}

	// Rule 3: Total is a multiple of 0.25
	if math.Abs(math.Mod(data.Total, 0.25)) < floatEpsilon || math.Abs(math.Mod(data.Total, 0.25)-0.25) < floatEpsilon {
		points += 25
	}

	// Rule 4: 5 points per two items
	points += int64(data.OriginalItems/2) * 5

	// Rule 5: Item description length multiple of 3
	for _, item := range data.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		if len(trimmedDesc) > 0 && len(trimmedDesc)%3 == 0 {
			points += int64(math.Ceil(item.Price * 0.2))
		}
	}

	// Rule 6: Odd purchase day
	if data.PurchaseDate.Day()%2 != 0 {
		points += 6
	}

	// Rule 7: Purchase time between 14:00 and 16:00 (exclusive interval)
	timeInMinutes := data.PurchaseTime.Hour()*60 + data.PurchaseTime.Minute()
	if timeInMinutes > 840 && timeInMinutes < 960 {
		points += 10
	}

	return points
}
