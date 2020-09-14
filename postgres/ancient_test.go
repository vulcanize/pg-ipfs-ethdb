// VulcanizeDB
// Copyright © 2019 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package pgipfsethdb_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/ipfs-ethdb/postgres"
)

var (
	ancientDB                ethdb.Database
	testBlockNumber          uint64 = 1
	testAncientHeader               = types.Header{Number: big.NewInt(2)}
	testAncientHeaderRLP, _         = rlp.EncodeToBytes(testHeader2)
	testAncientHash                 = testAncientHeader.Hash().Bytes()
	testAncientBodyBytes            = make([]byte, 10000)
	testAncientReceiptsBytes        = make([]byte, 5000)
	testAncientTD, _                = new(big.Int).SetString("1000000000000000000000", 10)
	testAncientTDBytes              = testAncientTD.Bytes()
)

var _ = Describe("Ancient", func() {
	BeforeEach(func() {
		db, err = pgipfsethdb.TestDB()
		Expect(err).ToNot(HaveOccurred())
		ancientDB = pgipfsethdb.NewDatabase(db)

	})
	AfterEach(func() {
		err = pgipfsethdb.ResetTestDB(db)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("AppendAncient/Sync/Has", func() {
		It("adds eth objects to the Ancient database and returns whether or not an ancient record exists", func() {
			hasAncient(testBlockNumber, false)

			err = ancientDB.AppendAncient(testBlockNumber, testAncientHash, testAncientHeaderRLP, testAncientBodyBytes, testAncientReceiptsBytes, testAncientTDBytes)
			Expect(err).ToNot(HaveOccurred())

			hasAncient(testBlockNumber, false)

			err = ancientDB.Sync()
			Expect(err).ToNot(HaveOccurred())

			hasAncient(testBlockNumber, true)
		})
	})

	Describe("AppendAncient/Sync/Ancient", func() {
		It("adds the eth objects to the Ancient database and returns the ancient objects on request", func() {
			hasAncient(testBlockNumber, false)

			_, err := ancientDB.Ancient(pgipfsethdb.FreezerHeaderTable, testBlockNumber)
			Expect(err).To(HaveOccurred())
			_, err = ancientDB.Ancient(pgipfsethdb.FreezerHashTable, testBlockNumber)
			Expect(err).To(HaveOccurred())
			_, err = ancientDB.Ancient(pgipfsethdb.FreezerBodiesTable, testBlockNumber)
			Expect(err).To(HaveOccurred())
			_, err = ancientDB.Ancient(pgipfsethdb.FreezerReceiptTable, testBlockNumber)
			Expect(err).To(HaveOccurred())
			_, err = ancientDB.Ancient(pgipfsethdb.FreezerDifficultyTable, testBlockNumber)
			Expect(err).To(HaveOccurred())

			err = ancientDB.AppendAncient(testBlockNumber, testAncientHash, testAncientHeaderRLP, testAncientBodyBytes, testAncientReceiptsBytes, testAncientTDBytes)
			Expect(err).ToNot(HaveOccurred())
			err = ancientDB.Sync()
			Expect(err).ToNot(HaveOccurred())

			hasAncient(testBlockNumber, true)

			ancientHeader, err := ancientDB.Ancient(pgipfsethdb.FreezerHeaderTable, testBlockNumber)
			Expect(err).ToNot(HaveOccurred())
			Expect(ancientHeader).To(Equal(testAncientHeaderRLP))

			ancientHash, err := ancientDB.Ancient(pgipfsethdb.FreezerHashTable, testBlockNumber)
			Expect(err).ToNot(HaveOccurred())
			Expect(ancientHash).To(Equal(testAncientHash))

			ancientBody, err := ancientDB.Ancient(pgipfsethdb.FreezerBodiesTable, testBlockNumber)
			Expect(err).ToNot(HaveOccurred())
			Expect(ancientBody).To(Equal(testAncientBodyBytes))

			ancientReceipts, err := ancientDB.Ancient(pgipfsethdb.FreezerReceiptTable, testBlockNumber)
			Expect(err).ToNot(HaveOccurred())
			Expect(ancientReceipts).To(Equal(testAncientReceiptsBytes))

			ancientTD, err := ancientDB.Ancient(pgipfsethdb.FreezerDifficultyTable, testBlockNumber)
			Expect(err).ToNot(HaveOccurred())
			Expect(ancientTD).To(Equal(testAncientTDBytes))
		})
	})

	Describe("AppendAncient/Sync/Ancients", func() {
		It("returns the height of the ancient database", func() {
			ancients, err := ancientDB.Ancients()
			Expect(err).ToNot(HaveOccurred())
			Expect(ancients).To(Equal(uint64(0)))

			for i := uint64(0); i <= 100; i++ {
				hasAncient(i, false)
				err = ancientDB.AppendAncient(i, testAncientHash, testAncientHeaderRLP, testAncientBodyBytes, testAncientReceiptsBytes, testAncientTDBytes)
				Expect(err).ToNot(HaveOccurred())
			}

			err = ancientDB.Sync()
			Expect(err).ToNot(HaveOccurred())

			for i := uint64(0); i <= 100; i++ {
				hasAncient(i, true)
			}
			ancients, err = ancientDB.Ancients()
			Expect(err).ToNot(HaveOccurred())
			Expect(ancients).To(Equal(uint64(100)))
		})
	})

	Describe("AppendAncient/Truncate/Sync", func() {
		It("truncates the ancient database to the provided height", func() {
			for i := uint64(0); i <= 100; i++ {
				hasAncient(i, false)
				err = ancientDB.AppendAncient(i, testAncientHash, testAncientHeaderRLP, testAncientBodyBytes, testAncientReceiptsBytes, testAncientTDBytes)
				Expect(err).ToNot(HaveOccurred())
			}

			err = ancientDB.Sync()
			Expect(err).ToNot(HaveOccurred())

			err = ancientDB.TruncateAncients(50)
			Expect(err).ToNot(HaveOccurred())

			for i := uint64(0); i <= 100; i++ {
				hasAncient(i, true)
			}

			ancients, err := ancientDB.Ancients()
			Expect(err).ToNot(HaveOccurred())
			Expect(ancients).To(Equal(uint64(100)))

			err = ancientDB.Sync()
			Expect(err).ToNot(HaveOccurred())

			for i := uint64(0); i <= 100; i++ {
				if i <= 50 {
					hasAncient(i, true)
				} else {
					hasAncient(i, false)
				}
			}

			ancients, err = ancientDB.Ancients()
			Expect(err).ToNot(HaveOccurred())
			Expect(ancients).To(Equal(uint64(50)))
		})
	})

	Describe("AppendAncient/Sync/AncientSize", func() {
		It("adds the eth objects to the Ancient database and returns the ancient objects on request", func() {
			for i := uint64(0); i <= 100; i++ {
				hasAncient(i, false)
				err = ancientDB.AppendAncient(i, testAncientHash, testAncientHeaderRLP, testAncientBodyBytes, testAncientReceiptsBytes, testAncientTDBytes)
				Expect(err).ToNot(HaveOccurred())
			}

			err = ancientDB.Sync()
			Expect(err).ToNot(HaveOccurred())

			for i := uint64(0); i <= 100; i++ {
				hasAncient(i, true)
			}

			ancientHeaderSize, err := ancientDB.AncientSize(pgipfsethdb.FreezerHeaderTable)
			Expect(err).ToNot(HaveOccurred())
			Expect(ancientHeaderSize).To(Equal(uint64(106496)))

			ancientHashSize, err := ancientDB.AncientSize(pgipfsethdb.FreezerHashTable)
			Expect(err).ToNot(HaveOccurred())
			Expect(ancientHashSize).To(Equal(uint64(32768)))

			ancientBodySize, err := ancientDB.AncientSize(pgipfsethdb.FreezerBodiesTable)
			Expect(err).ToNot(HaveOccurred())
			Expect(ancientBodySize).To(Equal(uint64(73728)))

			ancientReceiptsSize, err := ancientDB.AncientSize(pgipfsethdb.FreezerReceiptTable)
			Expect(err).ToNot(HaveOccurred())
			Expect(ancientReceiptsSize).To(Equal(uint64(65536)))

			ancientTDSize, err := ancientDB.AncientSize(pgipfsethdb.FreezerDifficultyTable)
			Expect(err).ToNot(HaveOccurred())
			Expect(ancientTDSize).To(Equal(uint64(32768)))
		})
	})
})

func hasAncient(blockNumber uint64, shouldHave bool) {
	has, err := ancientDB.HasAncient(pgipfsethdb.FreezerHeaderTable, blockNumber)
	Expect(err).ToNot(HaveOccurred())
	Expect(has).To(Equal(shouldHave))
	has, err = ancientDB.HasAncient(pgipfsethdb.FreezerHashTable, blockNumber)
	Expect(err).ToNot(HaveOccurred())
	Expect(has).To(Equal(shouldHave))
	has, err = ancientDB.HasAncient(pgipfsethdb.FreezerBodiesTable, blockNumber)
	Expect(err).ToNot(HaveOccurred())
	Expect(has).To(Equal(shouldHave))
	has, err = ancientDB.HasAncient(pgipfsethdb.FreezerReceiptTable, blockNumber)
	Expect(err).ToNot(HaveOccurred())
	Expect(has).To(Equal(shouldHave))
	has, err = ancientDB.HasAncient(pgipfsethdb.FreezerDifficultyTable, blockNumber)
	Expect(err).ToNot(HaveOccurred())
	Expect(has).To(Equal(shouldHave))
}