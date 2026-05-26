package spdb_debit_test

import (
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/provider/spdb_debit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpdbDebitParse(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SpdbDebit Parse Suite")
}

var _ = Describe("SpdbDebit Order", func() {
	Describe("New", func() {
		It("should create a new SpdbDebit instance", func() {
			sd := spdb_debit.New()
			Expect(sd).NotTo(BeNil())
			Expect(sd.Orders).To(BeEmpty())
			Expect(sd.LineNum).To(Equal(0))
		})
	})

	Describe("SetCurrency", func() {
		It("should set currency", func() {
			sd := spdb_debit.New()
			sd.SetCurrency("USD")
			Expect(sd.Currency).To(Equal("USD"))
		})

		It("should not set empty currency", func() {
			sd := spdb_debit.New()
			sd.SetCurrency("")
			Expect(sd.Currency).To(Equal("CNY")) // default
		})
	})

	Describe("Translate", func() {
		Context("with non-existent file", func() {
			It("should return error", func() {
				sd := spdb_debit.New()
				_, err := sd.Translate("non-existent-file.xls")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with unsupported file format", func() {
			It("should return error for CSV", func() {
				sd := spdb_debit.New()
				_, err := sd.Translate("test.csv")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported file format"))
			})
		})
	})
})