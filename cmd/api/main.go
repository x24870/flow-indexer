package main

import (
	"fmt"
)

func main() {
	borrow := 1500000.0
	month := 84
	// rate := 0.0238 // 2.38%
	yeild := 0.150 // 15%
	usdPrice := 31.0
	repay := 19404.0

	usdBalance := borrow / usdPrice
	fmt.Printf("TWD: %v -> USD: %v\n", int(borrow), usdBalance)
	fmt.Printf("Bitfinex yeild: %v\n", yeild)
	fmt.Printf("repay monthly: %v\n", repay)

	totalRepayTWD := 0.0

	for i := 1; i <= month; i++ {
		interest := getMonthlyInterest(usdBalance, yeild)
		usdBalance += interest
		repayUSD := getMonthlyRepay(repay, 27.0)
		usdBalance -= repayUSD
		totalRepayTWD += repay
		fmt.Printf("month: %d, interest: %f, repay: %f, usdBalance: %f\n", i, interest, repayUSD, usdBalance)
	}

}

func getMonthlyInterest(balance, yeild float64) float64 {
	return balance * yeild / 12
}

func getMonthlyRepay(repay, usdPrice float64) float64 {
	return repay / usdPrice
}
