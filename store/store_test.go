package store

import (
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jinzhu/gorm"
	"github.com/maichain/mapi/base/test"
	"github.com/maichain/mapi/types/reflect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager Test", func() {
	var (
		mysql *test.MySQLContainer
		db    *gorm.DB
	)
	BeforeSuite(func() {
		var err error
		mysql, err = test.NewMySQLContainer("quay.io/amis/eth-indexer-db-migration")
		Expect(mysql).ShouldNot(BeNil())
		Expect(err).Should(Succeed())
		Expect(mysql.Start()).Should(Succeed())

		db, err = gorm.Open("mysql", mysql.URL)
		Expect(err).Should(Succeed())
		Expect(db).ShouldNot(BeNil())

		db.LogMode(os.Getenv("ENABLE_DB_LOG_IN_TEST") != "")
	})

	AfterSuite(func() {
		mysql.Stop()
	})

	Context("Insert()", func() {
		It("should be ok", func() {
			manager := NewManager(db)
			header := &types.Header{
				Number: big.NewInt(10),
			}
			block := types.NewBlock(header, nil, []*types.Header{
				header,
			}, []*types.Receipt{
				types.NewReceipt([]byte{}, false, 0),
			})

			err := manager.InsertBlock(block, nil)
			Expect(err).Should(Succeed())

			By("insert the same block again, should be ok")
			err = manager.InsertBlock(block, nil)
			Expect(err).Should(Succeed())
		})

		It("failed due to wrong signer", func() {
			manager := NewManager(db)
			header := &types.Header{
				Number: big.NewInt(11),
			}
			block := types.NewBlock(header, []*types.Transaction{
				types.NewTransaction(0, common.Address{}, common.Big0, 0, common.Big0, []byte{}),
			}, []*types.Header{
				header,
			}, []*types.Receipt{
				types.NewReceipt([]byte{}, false, 0),
			})

			err := manager.InsertBlock(block, nil)
			Expect(err).Should(Equal(ErrWrongSigner))
		})
	})

	It("LatestHeader()", func() {
		manager := NewManager(db)
		block1 := types.NewBlockWithHeader(&types.Header{
			Number: big.NewInt(100),
		})
		block2 := types.NewBlockWithHeader(&types.Header{
			Number: big.NewInt(99),
		})
		err := manager.InsertBlock(block1, nil)
		Expect(err).Should(Succeed())
		err = manager.InsertBlock(block2, nil)
		Expect(err).Should(Succeed())

		header, err := manager.LatestHeader()
		Expect(err).Should(Succeed())
		Expect(reflect.DeepEqual(header, Header(block1))).Should(BeTrue())
	})
})

func TestBlockHeader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Store Test")
}