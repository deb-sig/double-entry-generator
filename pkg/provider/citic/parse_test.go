package citic_test

import (
	"testing"
	"time"

	"github.com/deb-sig/double-entry-generator/pkg/provider/citic"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCiticParse(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Citic Parse Suite")
}

var _ = Describe("Citic Parser", func() {
	var (
		citicParser *citic.Citic
	)

	BeforeEach(func() {
		citicParser = &citic.Citic{}
	})

	Describe("translateToOrders", func() {
		Context("with valid input data", func() {
			It("should successfully parse the order", func() {
				input := []string{
					"2023-01-01",       // Trade time
					"2023-01-02",       // Post time
					"Grocery Shopping", // Trade description
					"Credit Card",      // Method
					"",                 // Empty field (not used)
					"CNY",              // Currency
					"-100.50",          // Amount (negative for expense)
				}
				err := citicParser.TranslateToOrders(input)

				Expect(err).NotTo(HaveOccurred())
				Expect(citicParser.Orders).To(HaveLen(1))

				order := citicParser.Orders[0]
				expectedTradeTime, _ := time.Parse("2006-01-02 +0800 CST", "2023-01-01 +0800 CST")
				expectedPostTime, _ := time.Parse("2006-01-02 +0800 CST", "2023-01-02 +0800 CST")

				Expect(order.TradeTime).To(BeTemporally("==", expectedTradeTime))
				Expect(order.PostTime).To(BeTemporally("==", expectedPostTime))
				Expect(order.TradeDesc).To(Equal("Grocery Shopping"))
				Expect(order.Method).To(Equal("Credit Card"))
				Expect(order.Currency).To(Equal("CNY"))
				Expect(order.Type).To(Equal(citic.OrderType("收入")))
				Expect(order.Money).To(Equal(100.50))
			})
		})

		Context("with invalid trade time", func() {
			const invalidDate = "2023-01-02 10:00:00"
			It("should return an error", func() {
				input := []string{
					invalidDate,
					"2023-01-02 09:00:00",
					"Grocery Shopping",
					"Credit Card",
					"",
					"CNY",
					"-100.50",
				}

				err := citicParser.TranslateToOrders(input)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parse trade time"))
			})
		})

		Context("with invalid post time", func() {
			It("should return an error", func() {
				input := []string{
					"2023-01-01",
					"invalid-date",
					"Grocery Shopping",
					"Credit Card",
					"",
					"CNY",
					"-100.50",
				}

				err := citicParser.TranslateToOrders(input)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parse trade time"))
			})
		})

		Context("with invalid money format", func() {
			It("should return an error", func() {
				input := []string{
					"2023-12-08",
					"2023-12-08",
					"Grocery Shopping",
					"Credit Card",
					"",
					"CNY",
					"invalid-amount",
				}

				err := citicParser.TranslateToOrders(input)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parse money"))
			})
		})
	})
})
