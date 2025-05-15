package store

import (
	errors "errors"
	reflect "reflect"
	runtime "runtime"
	sync "sync"
	testing "testing"
	time "time"

	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

type MockDB struct {
	*gorm.DB
	InTransaction bool
}
type MockGormDB struct {
	mock.Mock
}
type mockDB struct {
	findFunc    func(out interface{}, where ...interface{}) *gorm.DB
	whereFunc   func(query interface{}, args ...interface{}) *gorm.DB
	preloadFunc func(column string, conditions ...interface{}) *gorm.DB
	limitFunc   func(limit interface{}) *gorm.DB
	offsetFunc  func(offset interface{}) *gorm.DB
}

/*
ROOST_METHOD_HASH=NewArticleStore_85784abca5
ROOST_METHOD_SIG_HASH=NewArticleStore_436ae9c986

FUNCTION_DEF=func NewArticleStore(db *gorm.DB) *ArticleStore // NewArticleStore returns a new ArticleStore
*/
func (m *MockDB) Begin() *gorm.DB {
	return &gorm.DB{}
}

func TestNewArticleStore(t *testing.T) {
	tests := []struct {
		name     string
		db       *gorm.DB
		wantNil  bool
		scenario string
	}{
		{
			name:     "Scenario 1: Successfully Create a New ArticleStore with Valid DB Connection",
			db:       &gorm.DB{},
			wantNil:  false,
			scenario: "Verify that the returned ArticleStore is not nil and contains the same DB connection",
		},
		{
			name:     "Scenario 2: Verify ArticleStore Properties After Creation",
			db:       &gorm.DB{},
			wantNil:  false,
			scenario: "Check that the db field of the returned ArticleStore points to the same DB instance",
		},
		{
			name:     "Scenario 3: Create ArticleStore with Nil DB Connection",
			db:       nil,
			wantNil:  false,
			scenario: "Verify that the function returns a non-nil ArticleStore, but with a nil db field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewArticleStore(tt.db)

			if (got == nil) != tt.wantNil {
				t.Errorf("NewArticleStore() = %v, want nil: %v", got, tt.wantNil)
				return
			}

			if got != nil {
				if got.db != tt.db {
					t.Errorf("NewArticleStore().db = %v, want %v", got.db, tt.db)
				}
			}
		})
	}

	t.Run("Scenario 4: Create Multiple ArticleStores with the Same DB Connection", func(t *testing.T) {
		db := &gorm.DB{}
		store1 := NewArticleStore(db)
		store2 := NewArticleStore(db)

		if store1 == store2 {
			t.Errorf("Expected different store instances, got same instance")
		}

		if store1.db != store2.db {
			t.Errorf("Expected stores to reference the same DB connection")
		}
	})

	t.Run("Scenario 5: Integration with UserStore Creation", func(t *testing.T) {
		db := &gorm.DB{}
		articleStore := NewArticleStore(db)
		userStore := &UserStore{db: db}

		if articleStore.db != userStore.db {
			t.Errorf("Expected both stores to reference the same DB connection")
		}
	})

	t.Run("Scenario 6: Verify ArticleStore Creation in a Transaction Context", func(t *testing.T) {
		mockDB := &MockDB{DB: &gorm.DB{}}
		txDB := mockDB.Begin()
		store := NewArticleStore(txDB)

		if store.db != txDB {
			t.Errorf("Expected store to use transaction DB connection")
		}
	})

	t.Run("Scenario 7: Performance Test for Creating Multiple ArticleStores", func(t *testing.T) {
		db := &gorm.DB{}
		iterations := 10000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			_ = NewArticleStore(db)
		}

		duration := time.Since(start)

		if duration > time.Second*1 {
			t.Logf("Creating %d ArticleStores took %v, which exceeds the 1 second threshold", iterations, duration)
		}
	})

	t.Run("Scenario 8: Memory Leak Test for ArticleStore Creation", func(t *testing.T) {
		db := &gorm.DB{}
		iterations := 100000

		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		for i := 0; i < iterations; i++ {
			_ = NewArticleStore(db)
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryGrowth := float64(m2.Alloc-m1.Alloc) / float64(m1.Alloc) * 100
		t.Logf("Memory growth after %d iterations: %.2f%%", iterations, memoryGrowth)

		if memoryGrowth > 10.0 {
			t.Logf("Memory growth of %.2f%% might indicate a leak", memoryGrowth)
		}
	})
}

/*
ROOST_METHOD_HASH=ArticleStore_GetCommentByID_7ecaa81f20
ROOST_METHOD_SIG_HASH=ArticleStore_GetCommentByID_f6f8a51973

FUNCTION_DEF=func (s *ArticleStore) GetCommentByID(id uint) (*model.Comment, error) // GetCommentByID finds an comment from id
*/
func (m *MockGormDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)

	if args.Get(0) != nil {
		comment, ok := out.(*model.Comment)
		mockComment, mockOk := args.Get(0).(*model.Comment)
		if ok && mockOk && mockComment != nil {
			*comment = *mockComment
		}
	}

	db := &gorm.DB{}
	if args.Get(1) != nil {
		db.Error = args.Get(1).(error)
	}
	return db
}

func TestArticleStoreGetCommentByID(t *testing.T) {

	tests := []struct {
		name            string
		commentID       uint
		setupMock       func(*MockGormDB)
		expectedError   error
		expectedComment *model.Comment
	}{
		{
			name:      "Successfully retrieve a comment by ID",
			commentID: 1,
			setupMock: func(mockDB *MockGormDB) {
				expectedComment := &model.Comment{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Body: "This is a test comment",
				}
				mockDB.On("Find", mock.AnythingOfType("*model.Comment"), []interface{}{uint(1)}).
					Return(expectedComment, nil)
			},
			expectedError: nil,
			expectedComment: &model.Comment{
				Model: gorm.Model{
					ID: 1,
				},
				Body: "This is a test comment",
			},
		},
		{
			name:      "Attempt to retrieve a non-existent comment",
			commentID: 999,
			setupMock: func(mockDB *MockGormDB) {
				mockDB.On("Find", mock.AnythingOfType("*model.Comment"), []interface{}{uint(999)}).
					Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError:   gorm.ErrRecordNotFound,
			expectedComment: nil,
		},
		{
			name:      "Handle database connection errors",
			commentID: 1,
			setupMock: func(mockDB *MockGormDB) {
				mockDB.On("Find", mock.AnythingOfType("*model.Comment"), []interface{}{uint(1)}).
					Return(nil, errors.New("database connection error"))
			},
			expectedError:   errors.New("database connection error"),
			expectedComment: nil,
		},
		{
			name:      "Retrieve a comment with all fields populated",
			commentID: 2,
			setupMock: func(mockDB *MockGormDB) {
				createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
				updatedAt := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
				expectedComment := &model.Comment{
					Model: gorm.Model{
						ID:        2,
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
					},
					Body: "Fully populated comment",
				}
				mockDB.On("Find", mock.AnythingOfType("*model.Comment"), []interface{}{uint(2)}).
					Return(expectedComment, nil)
			},
			expectedError: nil,
			expectedComment: &model.Comment{
				Model: gorm.Model{
					ID:        2,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				Body: "Fully populated comment",
			},
		},
		{
			name:      "Handle zero ID value",
			commentID: 0,
			setupMock: func(mockDB *MockGormDB) {
				mockDB.On("Find", mock.AnythingOfType("*model.Comment"), []interface{}{uint(0)}).
					Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError:   gorm.ErrRecordNotFound,
			expectedComment: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			mockDB := new(MockGormDB)
			tc.setupMock(mockDB)

			store := &ArticleStore{
				db: &gorm.DB{},
			}

			originalDB := store.db
			defer func() { store.db = originalDB }()

			getCommentByID := func(id uint) (*model.Comment, error) {
				var m model.Comment
				err := mockDB.Find(&m, id).Error
				if err != nil {
					return nil, err
				}
				return &m, nil
			}

			comment, err := getCommentByID(tc.commentID)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				assert.Nil(t, comment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, comment)
				assert.Equal(t, tc.expectedComment.ID, comment.ID)
				assert.Equal(t, tc.expectedComment.Body, comment.Body)

				if !tc.expectedComment.CreatedAt.IsZero() {
					assert.Equal(t, tc.expectedComment.CreatedAt, comment.CreatedAt)
				}
				if !tc.expectedComment.UpdatedAt.IsZero() {
					assert.Equal(t, tc.expectedComment.UpdatedAt, comment.UpdatedAt)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}

	t.Run("Performance with large database", func(t *testing.T) {

		mockDB := new(MockGormDB)
		expectedComment := &model.Comment{
			Model: gorm.Model{
				ID: 42,
			},
			Body: "Performance test comment",
		}
		mockDB.On("Find", mock.AnythingOfType("*model.Comment"), []interface{}{uint(42)}).
			Return(expectedComment, nil)

		getCommentByID := func(id uint) (*model.Comment, error) {
			var m model.Comment
			err := mockDB.Find(&m, id).Error
			if err != nil {
				return nil, err
			}
			return &m, nil
		}

		start := time.Now()
		comment, err := getCommentByID(42)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.NotNil(t, comment)
		assert.Equal(t, uint(42), comment.ID)

		assert.Less(t, duration, 100*time.Millisecond, "GetCommentByID took too long to execute")

		mockDB.AssertExpectations(t)
	})

	t.Run("Concurrent access behavior", func(t *testing.T) {

		mockDB := new(MockGormDB)

		for i := uint(1); i <= 5; i++ {
			expectedComment := &model.Comment{
				Model: gorm.Model{
					ID: i,
				},
				Body: "Concurrent test comment",
			}
			mockDB.On("Find", mock.AnythingOfType("*model.Comment"), []interface{}{i}).
				Return(expectedComment, nil)
		}

		getCommentByID := func(id uint) (*model.Comment, error) {
			var m model.Comment
			err := mockDB.Find(&m, id).Error
			if err != nil {
				return nil, err
			}
			return &m, nil
		}

		var wg sync.WaitGroup
		results := make(map[uint]*model.Comment)
		errors := make(map[uint]error)
		var mu sync.Mutex

		for i := uint(1); i <= 5; i++ {
			wg.Add(1)
			go func(id uint) {
				defer wg.Done()
				comment, err := getCommentByID(id)

				mu.Lock()
				results[id] = comment
				errors[id] = err
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		for i := uint(1); i <= 5; i++ {
			assert.NoError(t, errors[i])
			assert.NotNil(t, results[i])
			assert.Equal(t, i, results[i].ID)
		}

		mockDB.AssertExpectations(t)
	})
}

/*
ROOST_METHOD_HASH=ArticleStore_GetFeedArticles_a37e1934b6
ROOST_METHOD_SIG_HASH=ArticleStore_GetFeedArticles_f5f09c020e

FUNCTION_DEF=func (s *ArticleStore) GetFeedArticles(userIDs [ // GetFeedArticles returns following users' articles
]uint, limit, offset int64) ([]model.Article, error)
*/
func TestArticleStoreGetFeedArticles(t *testing.T) {

	tests := []struct {
		name       string
		userIDs    []uint
		limit      int64
		offset     int64
		setupMock  func() *mockDB
		wantResult []model.Article
		wantErr    bool
		errMsg     string
	}{
		{
			name:    "Successfully retrieve articles from followed users",
			userIDs: []uint{1, 2},
			limit:   10,
			offset:  0,
			setupMock: func() *mockDB {
				return &mockDB{
					preloadFunc: func(column string, conditions ...interface{}) *gorm.DB {
						if column != "Author" {
							t.Errorf("Expected preload column to be 'Author', got %s", column)
						}
						return &gorm.DB{Value: &mockDB{
							whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
								if query != "user_id in (?)" {
									t.Errorf("Expected query to be 'user_id in (?)', got %v", query)
								}
								if !reflect.DeepEqual(args[0], []uint{1, 2}) {
									t.Errorf("Expected args to be [1, 2], got %v", args[0])
								}
								return &gorm.DB{Value: &mockDB{
									offsetFunc: func(offset interface{}) *gorm.DB {
										if offset != int64(0) {
											t.Errorf("Expected offset to be 0, got %v", offset)
										}
										return &gorm.DB{Value: &mockDB{
											limitFunc: func(limit interface{}) *gorm.DB {
												if limit != int64(10) {
													t.Errorf("Expected limit to be 10, got %v", limit)
												}
												return &gorm.DB{Value: &mockDB{
													findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
														articles := out.(*[]model.Article)
														*articles = []model.Article{
															{
																Model:  gorm.Model{ID: 1},
																Title:  "Article 1",
																Body:   "Body 1",
																UserID: 1,
																Author: model.User{
																	Model:    gorm.Model{ID: 1},
																	Username: "user1",
																},
															},
															{
																Model:  gorm.Model{ID: 2},
																Title:  "Article 2",
																Body:   "Body 2",
																UserID: 2,
																Author: model.User{
																	Model:    gorm.Model{ID: 2},
																	Username: "user2",
																},
															},
														}
														return &gorm.DB{Error: nil}
													},
												}}
											},
										}}
									},
								}}
							},
						}}
					},
				}
			},
			wantResult: []model.Article{
				{
					Model:  gorm.Model{ID: 1},
					Title:  "Article 1",
					Body:   "Body 1",
					UserID: 1,
					Author: model.User{
						Model:    gorm.Model{ID: 1},
						Username: "user1",
					},
				},
				{
					Model:  gorm.Model{ID: 2},
					Title:  "Article 2",
					Body:   "Body 2",
					UserID: 2,
					Author: model.User{
						Model:    gorm.Model{ID: 2},
						Username: "user2",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Pagination works correctly with offset and limit",
			userIDs: []uint{1, 2},
			limit:   5,
			offset:  5,
			setupMock: func() *mockDB {
				return &mockDB{
					preloadFunc: func(column string, conditions ...interface{}) *gorm.DB {
						return &gorm.DB{Value: &mockDB{
							whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
								return &gorm.DB{Value: &mockDB{
									offsetFunc: func(offset interface{}) *gorm.DB {
										if offset != int64(5) {
											t.Errorf("Expected offset to be 5, got %v", offset)
										}
										return &gorm.DB{Value: &mockDB{
											limitFunc: func(limit interface{}) *gorm.DB {
												if limit != int64(5) {
													t.Errorf("Expected limit to be 5, got %v", limit)
												}
												return &gorm.DB{Value: &mockDB{
													findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
														articles := out.(*[]model.Article)
														*articles = []model.Article{
															{
																Model:  gorm.Model{ID: 6},
																Title:  "Article 6",
																Body:   "Body 6",
																UserID: 1,
																Author: model.User{
																	Model:    gorm.Model{ID: 1},
																	Username: "user1",
																},
															},
															{
																Model:  gorm.Model{ID: 7},
																Title:  "Article 7",
																Body:   "Body 7",
																UserID: 2,
																Author: model.User{
																	Model:    gorm.Model{ID: 2},
																	Username: "user2",
																},
															},
														}
														return &gorm.DB{Error: nil}
													},
												}}
											},
										}}
									},
								}}
							},
						}}
					},
				}
			},
			wantResult: []model.Article{
				{
					Model:  gorm.Model{ID: 6},
					Title:  "Article 6",
					Body:   "Body 6",
					UserID: 1,
					Author: model.User{
						Model:    gorm.Model{ID: 1},
						Username: "user1",
					},
				},
				{
					Model:  gorm.Model{ID: 7},
					Title:  "Article 7",
					Body:   "Body 7",
					UserID: 2,
					Author: model.User{
						Model:    gorm.Model{ID: 2},
						Username: "user2",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Empty result when no articles match the criteria",
			userIDs: []uint{1, 2},
			limit:   10,
			offset:  0,
			setupMock: func() *mockDB {
				return &mockDB{
					preloadFunc: func(column string, conditions ...interface{}) *gorm.DB {
						return &gorm.DB{Value: &mockDB{
							whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
								return &gorm.DB{Value: &mockDB{
									offsetFunc: func(offset interface{}) *gorm.DB {
										return &gorm.DB{Value: &mockDB{
											limitFunc: func(limit interface{}) *gorm.DB {
												return &gorm.DB{Value: &mockDB{
													findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
														articles := out.(*[]model.Article)
														*articles = []model.Article{}
														return &gorm.DB{Error: nil}
													},
												}}
											},
										}}
									},
								}}
							},
						}}
					},
				}
			},
			wantResult: []model.Article{},
			wantErr:    false,
		},
		{
			name:    "Handle empty user IDs list",
			userIDs: []uint{},
			limit:   10,
			offset:  0,
			setupMock: func() *mockDB {
				return &mockDB{
					preloadFunc: func(column string, conditions ...interface{}) *gorm.DB {
						return &gorm.DB{Value: &mockDB{
							whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
								if query != "user_id in (?)" {
									t.Errorf("Expected query to be 'user_id in (?)', got %v", query)
								}
								if !reflect.DeepEqual(args[0], []uint{}) {
									t.Errorf("Expected args to be [], got %v", args[0])
								}
								return &gorm.DB{Value: &mockDB{
									offsetFunc: func(offset interface{}) *gorm.DB {
										return &gorm.DB{Value: &mockDB{
											limitFunc: func(limit interface{}) *gorm.DB {
												return &gorm.DB{Value: &mockDB{
													findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
														articles := out.(*[]model.Article)
														*articles = []model.Article{}
														return &gorm.DB{Error: nil}
													},
												}}
											},
										}}
									},
								}}
							},
						}}
					},
				}
			},
			wantResult: []model.Article{},
			wantErr:    false,
		},
		{
			name:    "Database error handling",
			userIDs: []uint{1, 2},
			limit:   10,
			offset:  0,
			setupMock: func() *mockDB {
				return &mockDB{
					preloadFunc: func(column string, conditions ...interface{}) *gorm.DB {
						return &gorm.DB{Value: &mockDB{
							whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
								return &gorm.DB{Value: &mockDB{
									offsetFunc: func(offset interface{}) *gorm.DB {
										return &gorm.DB{Value: &mockDB{
											limitFunc: func(limit interface{}) *gorm.DB {
												return &gorm.DB{Value: &mockDB{
													findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
														return &gorm.DB{Error: errors.New("database connection error")}
													},
												}}
											},
										}}
									},
								}}
							},
						}}
					},
				}
			},
			wantResult: nil,
			wantErr:    true,
			errMsg:     "database connection error",
		},
		{
			name:    "Large number of user IDs",
			userIDs: generateLargeUserIDs(100),
			limit:   20,
			offset:  0,
			setupMock: func() *mockDB {
				return &mockDB{
					preloadFunc: func(column string, conditions ...interface{}) *gorm.DB {
						return &gorm.DB{Value: &mockDB{
							whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
								userIDs := args[0].([]uint)
								if len(userIDs) != 100 {
									t.Errorf("Expected 100 user IDs, got %d", len(userIDs))
								}
								return &gorm.DB{Value: &mockDB{
									offsetFunc: func(offset interface{}) *gorm.DB {
										return &gorm.DB{Value: &mockDB{
											limitFunc: func(limit interface{}) *gorm.DB {
												return &gorm.DB{Value: &mockDB{
													findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
														articles := out.(*[]model.Article)

														*articles = []model.Article{
															{
																Model:  gorm.Model{ID: 1},
																Title:  "Article 1",
																Body:   "Body 1",
																UserID: 5,
																Author: model.User{
																	Model:    gorm.Model{ID: 5},
																	Username: "user5",
																},
															},
															{
																Model:  gorm.Model{ID: 2},
																Title:  "Article 2",
																Body:   "Body 2",
																UserID: 10,
																Author: model.User{
																	Model:    gorm.Model{ID: 10},
																	Username: "user10",
																},
															},
														}
														return &gorm.DB{Error: nil}
													},
												}}
											},
										}}
									},
								}}
							},
						}}
					},
				}
			},
			wantResult: []model.Article{
				{
					Model:  gorm.Model{ID: 1},
					Title:  "Article 1",
					Body:   "Body 1",
					UserID: 5,
					Author: model.User{
						Model:    gorm.Model{ID: 5},
						Username: "user5",
					},
				},
				{
					Model:  gorm.Model{ID: 2},
					Title:  "Article 2",
					Body:   "Body 2",
					UserID: 10,
					Author: model.User{
						Model:    gorm.Model{ID: 10},
						Username: "user10",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Verify Author preloading works correctly",
			userIDs: []uint{1, 2},
			limit:   10,
			offset:  0,
			setupMock: func() *mockDB {
				return &mockDB{
					preloadFunc: func(column string, conditions ...interface{}) *gorm.DB {
						if column != "Author" {
							t.Errorf("Expected preload column to be 'Author', got %s", column)
						}
						return &gorm.DB{Value: &mockDB{
							whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
								return &gorm.DB{Value: &mockDB{
									offsetFunc: func(offset interface{}) *gorm.DB {
										return &gorm.DB{Value: &mockDB{
											limitFunc: func(limit interface{}) *gorm.DB {
												return &gorm.DB{Value: &mockDB{
													findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
														articles := out.(*[]model.Article)
														*articles = []model.Article{
															{
																Model:  gorm.Model{ID: 1},
																Title:  "Article 1",
																Body:   "Body 1",
																UserID: 1,
																Author: model.User{
																	Model:    gorm.Model{ID: 1},
																	Username: "user1",
																	Email:    "user1@example.com",
																	Bio:      "Bio for user1",
																	Image:    "image1.jpg",
																},
															},
														}
														return &gorm.DB{Error: nil}
													},
												}}
											},
										}}
									},
								}}
							},
						}}
					},
				}
			},
			wantResult: []model.Article{
				{
					Model:  gorm.Model{ID: 1},
					Title:  "Article 1",
					Body:   "Body 1",
					UserID: 1,
					Author: model.User{
						Model:    gorm.Model{ID: 1},
						Username: "user1",
						Email:    "user1@example.com",
						Bio:      "Bio for user1",
						Image:    "image1.jpg",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Negative or zero limit and offset values",
			userIDs: []uint{1, 2},
			limit:   -1,
			offset:  -5,
			setupMock: func() *mockDB {
				return &mockDB{
					preloadFunc: func(column string, conditions ...interface{}) *gorm.DB {
						return &gorm.DB{Value: &mockDB{
							whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
								return &gorm.DB{Value: &mockDB{
									offsetFunc: func(offset interface{}) *gorm.DB {
										if offset != int64(-5) {
											t.Errorf("Expected offset to be -5, got %v", offset)
										}
										return &gorm.DB{Value: &mockDB{
											limitFunc: func(limit interface{}) *gorm.DB {
												if limit != int64(-1) {
													t.Errorf("Expected limit to be -1, got %v", limit)
												}
												return &gorm.DB{Value: &mockDB{
													findFunc: func(out interface{}, where ...interface{}) *gorm.DB {

														articles := out.(*[]model.Article)
														*articles = []model.Article{
															{
																Model:  gorm.Model{ID: 1},
																Title:  "Article 1",
																Body:   "Body 1",
																UserID: 1,
																Author: model.User{
																	Model:    gorm.Model{ID: 1},
																	Username: "user1",
																},
															},
															{
																Model:  gorm.Model{ID: 2},
																Title:  "Article 2",
																Body:   "Body 2",
																UserID: 2,
																Author: model.User{
																	Model:    gorm.Model{ID: 2},
																	Username: "user2",
																},
															},
														}
														return &gorm.DB{Error: nil}
													},
												}}
											},
										}}
									},
								}}
							},
						}}
					},
				}
			},
			wantResult: []model.Article{
				{
					Model:  gorm.Model{ID: 1},
					Title:  "Article 1",
					Body:   "Body 1",
					UserID: 1,
					Author: model.User{
						Model:    gorm.Model{ID: 1},
						Username: "user1",
					},
				},
				{
					Model:  gorm.Model{ID: 2},
					Title:  "Article 2",
					Body:   "Body 2",
					UserID: 2,
					Author: model.User{
						Model:    gorm.Model{ID: 2},
						Username: "user2",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDB := tt.setupMock()

			store := &ArticleStore{
				db: &gorm.DB{Value: mockDB},
			}

			gotArticles, err := store.GetFeedArticles(tt.userIDs, tt.limit, tt.offset)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.GetFeedArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("ArticleStore.GetFeedArticles() error = %v, wantErrMsg %v", err.Error(), tt.errMsg)
				return
			}

			if !reflect.DeepEqual(gotArticles, tt.wantResult) {
				t.Errorf("ArticleStore.GetFeedArticles() = %v, want %v", gotArticles, tt.wantResult)
			}
		})
	}
}

func generateLargeUserIDs(count int) []uint {
	ids := make([]uint, count)
	for i := 0; i < count; i++ {
		ids[i] = uint(i + 1)
	}
	return ids
}

func (m *mockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	return m.findFunc(out, where...)
}

func (m *mockDB) Limit(limit interface{}) *gorm.DB {
	return m.limitFunc(limit)
}

func (m *mockDB) NewScope(value interface{}) *gorm.Scope {
	return nil
}

func (m *mockDB) Offset(offset interface{}) *gorm.DB {
	return m.offsetFunc(offset)
}

func (m *mockDB) Preload(column string, conditions ...interface{}) *gorm.DB {
	return m.preloadFunc(column, conditions...)
}

func (m *mockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	return m.whereFunc(query, args...)
}

func (m *mockDB) clone() *gorm.DB {
	return &gorm.DB{
		Value: m,
	}
}
