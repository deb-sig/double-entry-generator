package spdb_debit_test

import (
	"testing"

	main "github.com/deb-sig/double-entry-generator/v2/pkg/analyser/spdb_debit"
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	spdbConfig "github.com/deb-sig/double-entry-generator/v2/pkg/provider/spdb_debit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpdbDebit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SpdbDebit Suite")
}

var _ = Describe("SpdbDebit Analyser", func() {
	var (
		analyser main.SpdbDebit
	)

	BeforeEach(func() {
		analyser = main.SpdbDebit{}
	})

	Describe("GetAllCandidateAccounts", func() {
		Context("when config is nil", func() {
			It("should return only default accounts", func() {
				cfg := &config.Config{
					DefaultPlusAccount:  "DefaultPlus",
					DefaultMinusAccount: "DefaultMinus",
					DefaultCashAccount:  "DefaultCash",
				}

				result := analyser.GetAllCandidateAccounts(cfg)

				Expect(result).To(HaveLen(3))
				Expect(result).To(HaveKey("DefaultPlus"))
				Expect(result).To(HaveKey("DefaultMinus"))
				Expect(result).To(HaveKey("DefaultCash"))
			})
		})

		Context("when config has rules with accounts", func() {
			It("should return all accounts from rules plus default accounts", func() {
				targetAcc1 := "Expenses:Food"
				targetAcc2 := "Expenses:Transport"

				cfg := &config.Config{
					DefaultPlusAccount:  "DefaultPlus",
					DefaultMinusAccount: "DefaultMinus",
					DefaultCashAccount:  "DefaultCash",
					SpdbDebit: &spdbConfig.Config{
						Rules: []spdbConfig.Rule{
							{TargetAccount: &targetAcc1},
							{TargetAccount: &targetAcc2},
						},
					},
				}

				result := analyser.GetAllCandidateAccounts(cfg)

				Expect(result).To(HaveLen(5))
				Expect(result).To(HaveKey("Expenses:Food"))
				Expect(result).To(HaveKey("Expenses:Transport"))
				Expect(result).To(HaveKey("DefaultPlus"))
				Expect(result).To(HaveKey("DefaultMinus"))
				Expect(result).To(HaveKey("DefaultCash"))
			})
		})
	})

	Describe("GetAccountsAndTags", func() {
		Context("when config is nil", func() {
			It("should return default accounts", func() {
				order := &ir.Order{
					Type: ir.TypeSend,
				}
				cfg := &config.Config{
					DefaultMinusAccount: "Assets:SPDB",
					DefaultPlusAccount:  "Expenses:FIXME",
				}

				ignore, minus, plus, accounts, tags := analyser.GetAccountsAndTags(order, cfg, "beancount", "spdb_debit")

				Expect(ignore).To(BeFalse())
				Expect(minus).To(Equal("Assets:SPDB"))
				Expect(plus).To(Equal("Expenses:FIXME"))
				Expect(accounts).To(BeNil())
				Expect(tags).To(BeEmpty())
			})
		})

		Context("with TxType matching", func() {
			It("should match order by TxType", func() {
				txTypeTransfer := "转账"
				txTypeConsume := "消费"
				transferAccount := "Expenses:Transfer"
				consumeAccount := "Expenses:Consumption"

				cfg := &config.Config{
					DefaultMinusAccount: "Assets:SPDB",
					DefaultPlusAccount:  "Expenses:FIXME",
					SpdbDebit: &spdbConfig.Config{
						Rules: []spdbConfig.Rule{
							{TxType: &txTypeTransfer, TargetAccount: &transferAccount},
							{TxType: &txTypeConsume, TargetAccount: &consumeAccount},
						},
					},
				}

				// Test transfer transaction
				orderTransfer := &ir.Order{
					Type:           ir.TypeSend,
					TxTypeOriginal: "转账",
					Peer:           "对方",
				}
				_, _, plus, _, _ := analyser.GetAccountsAndTags(orderTransfer, cfg, "beancount", "spdb_debit")
				Expect(plus).To(Equal("Expenses:Transfer"))

				// Test consume transaction
				orderConsume := &ir.Order{
					Type:           ir.TypeSend,
					TxTypeOriginal: "消费",
					Peer:           "商家",
				}
				_, _, plus, _, _ = analyser.GetAccountsAndTags(orderConsume, cfg, "beancount", "spdb_debit")
				Expect(plus).To(Equal("Expenses:Consumption"))
			})
		})

		Context("with Tags matching", func() {
			It("should return tags when rule matches", func() {
				peerExpense := "超市"
				targetAcc := "Expenses:Shopping"
				tagStr := "shopping,grocery"

				cfg := &config.Config{
					DefaultMinusAccount: "Assets:SPDB",
					DefaultPlusAccount:  "Expenses:FIXME",
					SpdbDebit: &spdbConfig.Config{
						Rules: []spdbConfig.Rule{
							{Peer: &peerExpense, TargetAccount: &targetAcc, Tags: &tagStr},
						},
					},
				}

				order := &ir.Order{
					Type: ir.TypeSend,
					Peer: "大润发超市",
				}
				_, _, _, _, tags := analyser.GetAccountsAndTags(order, cfg, "beancount", "spdb_debit")

				Expect(tags).To(HaveLen(2))
				Expect(tags).To(ContainElements("shopping", "grocery"))
			})
		})

		Context("with Peer matching", func() {
			It("should match order by peer", func() {
				peerSalary := "工资"
				incomeAccount := "Income:Salary"

				cfg := &config.Config{
					DefaultMinusAccount: "Assets:SPDB",
					DefaultPlusAccount:  "Expenses:FIXME",
					SpdbDebit: &spdbConfig.Config{
						Rules: []spdbConfig.Rule{
							{Peer: &peerSalary, TargetAccount: &incomeAccount},
						},
					},
				}

				order := &ir.Order{
					Type: ir.TypeRecv,
					Peer: "公司名称-工资",
				}
				_, minus, _, _, _ := analyser.GetAccountsAndTags(order, cfg, "beancount", "spdb_debit")
				Expect(minus).To(Equal("Income:Salary"))
			})
		})

		Context("with Item matching", func() {
			It("should match order by item", func() {
				itemFood := "餐饮"
				foodAccount := "Expenses:Food"

				cfg := &config.Config{
					DefaultMinusAccount: "Assets:SPDB",
					DefaultPlusAccount:  "Expenses:FIXME",
					SpdbDebit: &spdbConfig.Config{
						Rules: []spdbConfig.Rule{
							{Item: &itemFood, TargetAccount: &foodAccount},
						},
					},
				}

				order := &ir.Order{
					Type: ir.TypeSend,
					Item: "餐饮美食-午餐",
				}
				_, _, plus, _, _ := analyser.GetAccountsAndTags(order, cfg, "beancount", "spdb_debit")
				Expect(plus).To(Equal("Expenses:Food"))
			})
		})

		Context("with FullMatch enabled", func() {
			It("should only match exact peer", func() {
				peerExact := "超市"
				exactAccount := "Expenses:Grocery"

				cfg := &config.Config{
					DefaultMinusAccount: "Assets:SPDB",
					DefaultPlusAccount:  "Expenses:FIXME",
					SpdbDebit: &spdbConfig.Config{
						Rules: []spdbConfig.Rule{
							{Peer: &peerExact, TargetAccount: &exactAccount, FullMatch: true},
						},
					},
				}

				// Exact match should succeed
				orderExact := &ir.Order{
					Type: ir.TypeSend,
					Peer: "超市",
				}
				_, _, plus, _, _ := analyser.GetAccountsAndTags(orderExact, cfg, "beancount", "spdb_debit")
				Expect(plus).To(Equal("Expenses:Grocery"))

				// Partial match should fail
				orderPartial := &ir.Order{
					Type: ir.TypeSend,
					Peer: "大润发超市",
				}
				_, _, plus, _, _ = analyser.GetAccountsAndTags(orderPartial, cfg, "beancount", "spdb_debit")
				Expect(plus).To(Equal("Expenses:FIXME")) // Default account
			})
		})

		Context("with Ignore flag", func() {
			It("should ignore the order when rule matches", func() {
				peerIgnore := "测试"

				cfg := &config.Config{
					DefaultMinusAccount: "Assets:SPDB",
					DefaultPlusAccount:  "Expenses:FIXME",
					SpdbDebit: &spdbConfig.Config{
						Rules: []spdbConfig.Rule{
							{Peer: &peerIgnore, Ignore: true},
						},
					},
				}

				order := &ir.Order{
					Type: ir.TypeSend,
					Peer: "测试交易",
				}
				ignore, _, _, _, _ := analyser.GetAccountsAndTags(order, cfg, "beancount", "spdb_debit")
				Expect(ignore).To(BeTrue())
			})
		})

		Context("with custom separator", func() {
			It("should split tags with custom separator", func() {
				peer := "奖金"
				targetAcc := "Income:Bonus"
				tagStr := "bonus|year-end|extra"
				sep := "|"

				cfg := &config.Config{
					DefaultMinusAccount: "Assets:SPDB",
					DefaultPlusAccount:  "Expenses:FIXME",
					SpdbDebit: &spdbConfig.Config{
						Rules: []spdbConfig.Rule{
							{Peer: &peer, TargetAccount: &targetAcc, Tags: &tagStr, Separator: &sep},
						},
					},
				}

				order := &ir.Order{
					Type: ir.TypeRecv,
					Peer: "年终奖金",
				}
				_, _, _, _, tags := analyser.GetAccountsAndTags(order, cfg, "beancount", "spdb_debit")

				Expect(tags).To(HaveLen(3))
				Expect(tags).To(ContainElements("bonus", "year-end", "extra"))
			})
		})

		Context("with income order (TypeRecv)", func() {
			It("should set correct account direction for income", func() {
				peer := "工资"
				incomeAccount := "Income:Salary"

				cfg := &config.Config{
					DefaultMinusAccount: "Assets:SPDB",
					DefaultPlusAccount:  "Expenses:FIXME",
					SpdbDebit: &spdbConfig.Config{
						Rules: []spdbConfig.Rule{
							{Peer: &peer, TargetAccount: &incomeAccount},
						},
					},
				}

				order := &ir.Order{
					Type: ir.TypeRecv, // 收入
					Peer: "公司-工资",
				}
				_, minus, plus, _, _ := analyser.GetAccountsAndTags(order, cfg, "beancount", "spdb_debit")
				// 收入：资产账户(plus)增加，对方账户(minus)减少
				Expect(plus).To(Equal("Assets:SPDB"))
				Expect(minus).To(Equal("Income:Salary"))
			})
		})

		Context("with expense order (TypeSend)", func() {
			It("should set correct account direction for expense", func() {
				peer := "超市"
				expenseAccount := "Expenses:Shopping"

				cfg := &config.Config{
					DefaultMinusAccount: "Assets:SPDB",
					DefaultPlusAccount:  "Expenses:FIXME",
					SpdbDebit: &spdbConfig.Config{
						Rules: []spdbConfig.Rule{
							{Peer: &peer, TargetAccount: &expenseAccount},
						},
					},
				}

				order := &ir.Order{
					Type: ir.TypeSend, // 支出
					Peer: "大润发超市",
				}
				_, minus, plus, _, _ := analyser.GetAccountsAndTags(order, cfg, "beancount", "spdb_debit")
				// 支出：对方账户(plus)增加，资产账户(minus)减少
				Expect(plus).To(Equal("Expenses:Shopping"))
				Expect(minus).To(Equal("Assets:SPDB"))
			})
		})
	})
})