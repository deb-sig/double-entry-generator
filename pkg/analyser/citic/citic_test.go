package citic_test

import (
	main "github.com/deb-sig/double-entry-generator/pkg/analyser/citic"
	"github.com/deb-sig/double-entry-generator/pkg/config"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
	citicConfig "github.com/deb-sig/double-entry-generator/pkg/provider/citic"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

// Entry point for Ginkgo tests
func TestCitic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Citic Suite")
}

var _ = Describe("Citic", func() {
	var (
		citic main.Citic
	)

	BeforeEach(func() {
		citic = main.Citic{}
	})

	Describe("GetAllCandidateAccounts", func() {
		Context("when config is nil or has no rules", func() {
			It("should return only default accounts", func() {
				cfg := &config.Config{
					DefaultPlusAccount:  "DefaultPlus",
					DefaultMinusAccount: "DefaultMinus",
				}

				result := citic.GetAllCandidateAccounts(cfg)

				Expect(result).To(HaveLen(0))
			})
		})

		Context("when config has rules with accounts", func() {
			It("should return all accounts from rules plus default accounts", func() {
				methodAcc := "Account1"
				targetAcc := "Account2"
				methodAcc2 := "Account3"

				cfg := &config.Config{
					DefaultPlusAccount:  "DefaultPlus",
					DefaultMinusAccount: "DefaultMinus",
					Citic: &citicConfig.Config{
						Rules: []citicConfig.Rule{
							{
								MethodAccount: &methodAcc,
								TargetAccount: &targetAcc,
							},
							{
								MethodAccount: &methodAcc2,
								TargetAccount: nil,
							},
						},
					},
				}

				result := citic.GetAllCandidateAccounts(cfg)

				Expect(result).To(HaveLen(5))
				Expect(result).To(HaveKey("Account1"))
				Expect(result).To(HaveKey("Account2"))
				Expect(result).To(HaveKey("Account3"))
				Expect(result).To(HaveKey("DefaultPlus"))
				Expect(result).To(HaveKey("DefaultMinus"))
			})
		})
	})

	Describe("GetAccountsAndTags", func() {
		Context("when config is nil or has no rules", func() {
			It("should return default accounts", func() {
				order := &ir.Order{
					Type: ir.TypeSend,
				}
				cfg := &config.Config{
					DefaultMinusAccount: "DefaultMinus",
					DefaultPlusAccount:  "DefaultPlus",
				}

				ignore, minus, plus, accounts, tags := citic.GetAccountsAndTags(order, cfg, "target", "provider")

				Expect(ignore).To(BeFalse())
				Expect(minus).To(Equal("DefaultMinus"))
				Expect(plus).To(Equal("DefaultPlus"))
				Expect(accounts).To(BeNil())
				Expect(tags).To(BeNil())
			})
		})
	})
})
