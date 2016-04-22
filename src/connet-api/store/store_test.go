package store_test

import (
	"connet-api/fakes"
	"connet-api/models"
	"connet-api/store"
	"errors"
	"fmt"
	"lib/db"
	"lib/testsupport"
	"math/rand"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {
	var dataStore store.Store
	var testDatabase *testsupport.TestDatabase
	var realDb *sqlx.DB
	var mockDb *fakes.Db

	BeforeEach(func() {
		mockDb = &fakes.Db{}

		dbName := fmt.Sprintf("test_connet_database_%x", rand.Int())
		dbConnectionInfo := testsupport.GetDBConnectionInfo()
		testDatabase = dbConnectionInfo.CreateDatabase(dbName)

		var err error
		realDb, err = db.GetConnectionPool(testDatabase.URL())
		Expect(err).NotTo(HaveOccurred())
		dataStore, err = store.New(realDb)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if realDb != nil {
			Expect(realDb.Close()).To(Succeed())
		}
		if testDatabase != nil {
			testDatabase.Destroy()
		}
	})

	Describe("Connecting to the database and migrating", func() {
		It("returns a store", func() {
			dbStore, err := store.New(realDb)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbStore).NotTo(BeNil())
		})

		Context("when the tables already exist", func() {
			It("succeeds", func() {
				_, err := store.New(realDb)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the db operation fails", func() {
			BeforeEach(func() {
				mockDb.ExecReturns(nil, errors.New("some error"))
			})

			It("should return a sensible error", func() {
				_, err := store.New(mockDb)
				Expect(err).To(MatchError("setting up tables: some error"))
			})
		})
	})

	Describe("Create", func() {
		It("stores the route", func() {
			route := models.Route{
				AppGuid: "my-application-guid",
				Fqdn:    "my-application.cloudfoundry",
			}

			err := dataStore.Create(route)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the insert fails", func() {
			Context("and the error is a 'pq' error", func() {
				BeforeEach(func() {
					mockDb.NamedExecReturns(nil, &pq.Error{Code: "2201G"})
				})

				It("should return the error code", func() {
					store, err := store.New(mockDb)
					Expect(err).NotTo(HaveOccurred())

					err = store.Create(models.Route{})
					Expect(err).To(MatchError("insert: invalid_argument_for_width_bucket_function"))
				})
			})

			Context("and the failure is not a pq Error", func() {
				BeforeEach(func() {
					mockDb.NamedExecReturns(nil, errors.New("some-insert-error"))
				})

				It("should return a sensible error", func() {
					store, err := store.New(mockDb)
					Expect(err).NotTo(HaveOccurred())

					err = store.Create(models.Route{})
					Expect(err).To(MatchError("insert: some-insert-error"))
				})
			})
		})
	})

	Describe("All", func() {
		var routes []models.Route

		BeforeEach(func() {
			routes = []models.Route{
				models.Route{
					AppGuid: "my-application-guid",
					Fqdn:    "my-application.cloudfoundry",
				},
				models.Route{
					AppGuid: "another-application-guid",
					Fqdn:    "another-application.cloudfoundry",
				},
				models.Route{
					AppGuid: "and-another-application-guid",
					Fqdn:    "and-another-application.cloudfoundry",
				},
			}

			for _, route := range routes {
				err := dataStore.Create(route)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("retrieves all of the routes", func() {
			r, err := dataStore.All()
			Expect(err).NotTo(HaveOccurred())

			Expect(r).To(ConsistOf(routes))
		})

		Context("when the query fails", func() {
			Context("and the error is a 'pq' error", func() {
				BeforeEach(func() {
					mockDb.SelectReturns(&pq.Error{Code: "2201G"})
				})

				It("should return the error code", func() {
					store, err := store.New(mockDb)
					Expect(err).NotTo(HaveOccurred())

					_, err = store.All()
					Expect(err).To(MatchError("select: invalid_argument_for_width_bucket_function"))
				})
			})

			Context("and the failure is not a pq Error", func() {
				BeforeEach(func() {
					mockDb.SelectReturns(errors.New("some-select-error"))
				})

				It("should return a sensible error", func() {
					store, err := store.New(mockDb)
					Expect(err).NotTo(HaveOccurred())

					_, err = store.All()
					Expect(err).To(MatchError("select: some-select-error"))
				})
			})
		})
	})
})
