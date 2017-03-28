package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/almighty/almighty-core/account"
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/app/test"
	. "github.com/almighty/almighty-core/controller"
	"github.com/almighty/almighty-core/gormapplication"
	"github.com/almighty/almighty-core/gormsupport"
	"github.com/almighty/almighty-core/gormsupport/cleaner"
	"github.com/almighty/almighty-core/gormtestsupport"
	"github.com/almighty/almighty-core/resource"
	testsupport "github.com/almighty/almighty-core/test"
	almtoken "github.com/almighty/almighty-core/token"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

func TestUsers(t *testing.T) {
	resource.Require(t, resource.Database)
	suite.Run(t, &TestUsersSuite{DBTestSuite: gormtestsupport.NewDBTestSuite("../config.yaml")})
}

type TestUsersSuite struct {
	gormtestsupport.DBTestSuite
	db           *gormapplication.GormDB
	svc          *goa.Service
	clean        func()
	controller   *UsersController
	userRepo     account.UserRepository
	identityRepo account.IdentityRepository
}

func (s *TestUsersSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()
	s.clean = cleaner.DeleteCreatedEntities(s.DB)
	s.svc = goa.New("test")
	s.db = gormapplication.NewGormDB(s.DB)
	s.controller = NewUsersController(s.svc, s.db, s.Configuration)
	s.userRepo = s.db.Users()
	s.identityRepo = s.db.Identities()
	s.Configuration = wibConfiguration
}

func (s *TestUsersSuite) TearDownSuite() {
	s.clean()
}

func (s *TestUsersSuite) SecuredController(identity account.Identity) (*goa.Service, *UsersController) {
	priv, _ := almtoken.ParsePrivateKey([]byte(almtoken.RSAPrivateKey))

	svc := testsupport.ServiceAsUser("Status-Service", almtoken.NewManagerWithPrivateKey(priv), identity)
	return svc, NewUsersController(svc, s.db, s.Configuration)
}

func (s *TestUsersSuite) TestUpdateUserOK() {
	// given
	user := s.createRandomUser()
	identity := s.createRandomIdentity(user, account.KeycloakIDP)
	_, result := test.ShowUsersOK(s.T(), nil, nil, s.controller, identity.ID.String(), nil, nil)
	assert.Equal(s.T(), user.ID.String(), *result.Data.ID)
	assert.Equal(s.T(), user.FullName, *result.Data.Attributes.FullName)
	assert.Equal(s.T(), user.ImageURL, *result.Data.Attributes.ImageURL)
	assert.Equal(s.T(), identity.ProviderType, *result.Data.Attributes.ProviderType)
	assert.Equal(s.T(), identity.Username, *result.Data.Attributes.Username)
	// when
	newEmail := "TestUpdateUserOK-" + uuid.NewV4().String() + "@email.com"
	newFullName := "TestUpdateUserOK"
	newImageURL := "http://new.image.io/imageurl"
	newBio := "new bio"
	newProfileURL := "http://new.profile.url/url"
	secureService, secureController := s.SecuredController(identity)

	contextInformation := map[string]interface{}{
		"last_visited": "yesterday",
		"space":        "3d6dab8d-f204-42e8-ab29-cdb1c93130ad",
		"rate":         100.00,
		"count":        3,
	}
	//secureController, secureService := createSecureController(t, identity)
	updateUsersPayload := createUpdateUsersPayload(&newEmail, &newFullName, &newBio, &newImageURL, &newProfileURL, contextInformation)
	_, result = test.UpdateUsersOK(s.T(), secureService.Context, secureService, secureController, updateUsersPayload)
	// then
	require.NotNil(s.T(), result)
	// let's fetch it and validate
	_, result = test.ShowUsersOK(s.T(), nil, nil, s.controller, identity.ID.String(), nil, nil)
	require.NotNil(s.T(), result)
	assert.Equal(s.T(), user.ID.String(), *result.Data.ID)
	assert.Equal(s.T(), newFullName, *result.Data.Attributes.FullName)
	assert.Equal(s.T(), newImageURL, *result.Data.Attributes.ImageURL)
	assert.Equal(s.T(), newBio, *result.Data.Attributes.Bio)
	assert.Equal(s.T(), newProfileURL, *result.Data.Attributes.URL)
	updatedContextInformation := result.Data.Attributes.ContextInformation
	assert.Equal(s.T(), contextInformation["last_visited"], updatedContextInformation["last_visited"])
	countValue, ok := updatedContextInformation["count"].(float64)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), contextInformation["count"], int(countValue))
	assert.Equal(s.T(), contextInformation["rate"], updatedContextInformation["rate"])
}

/*
	Test to unset variable in contextInformation
*/

func (s *TestUsersSuite) TestUpdateUserUnsetVariableInContextInfo() {
	// given
	user := s.createRandomUser()
	identity := s.createRandomIdentity(user, account.KeycloakIDP)
	_, result := test.ShowUsersOK(s.T(), nil, nil, s.controller, identity.ID.String(), nil, nil)
	assert.Equal(s.T(), user.ID.String(), *result.Data.ID)
	assert.Equal(s.T(), user.FullName, *result.Data.Attributes.FullName)
	assert.Equal(s.T(), user.ImageURL, *result.Data.Attributes.ImageURL)
	assert.Equal(s.T(), identity.ProviderType, *result.Data.Attributes.ProviderType)
	assert.Equal(s.T(), identity.Username, *result.Data.Attributes.Username)
	// when
	newEmail := "TestUpdateUserUnsetVariableInContextInfo-" + uuid.NewV4().String() + "@email.com"
	newFullName := "TestUpdateUserUnsetVariableInContextInfo"
	newImageURL := "http://new.image.io/imageurl"
	newBio := "new bio"
	newProfileURL := "http://new.profile.url/url"
	secureService, secureController := s.SecuredController(identity)
	contextInformation := map[string]interface{}{
		"last_visited": "yesterday",
		"space":        "3d6dab8d-f204-42e8-ab29-cdb1c93130ad",
		"rate":         100.00,
		"count":        3,
	}
	//secureController, secureService := createSecureController(t, identity)
	updateUsersPayload := createUpdateUsersPayload(&newEmail, &newFullName, &newBio, &newImageURL, &newProfileURL, contextInformation)
	_, result = test.UpdateUsersOK(s.T(), secureService.Context, secureService, secureController, updateUsersPayload)
	// then
	require.NotNil(s.T(), result)
	// let's fetch it and validate the usual stuff.
	_, result = test.ShowUsersOK(s.T(), nil, nil, s.controller, identity.ID.String(), nil, nil)
	require.NotNil(s.T(), result)
	assert.Equal(s.T(), user.ID.String(), *result.Data.ID)
	assert.Equal(s.T(), newFullName, *result.Data.Attributes.FullName)
	assert.Equal(s.T(), newImageURL, *result.Data.Attributes.ImageURL)
	assert.Equal(s.T(), newBio, *result.Data.Attributes.Bio)
	assert.Equal(s.T(), newProfileURL, *result.Data.Attributes.URL)
	updatedContextInformation := result.Data.Attributes.ContextInformation
	assert.Equal(s.T(), contextInformation["last_visited"], updatedContextInformation["last_visited"])

	/** Usual stuff done, now lets unset **/
	contextInformation = map[string]interface{}{
		"last_visited": nil,
		"space":        "3d6dab8d-f204-42e8-ab29-cdb1c93130ad",
		"rate":         100.00,
		"count":        3,
	}

	updateUsersPayload = createUpdateUsersPayload(&newEmail, &newFullName, &newBio, &newImageURL, &newProfileURL, contextInformation)
	_, result = test.UpdateUsersOK(s.T(), secureService.Context, secureService, secureController, updateUsersPayload)
	// then
	require.NotNil(s.T(), result)
	// let's fetch it and validate the usual stuff.
	_, result = test.ShowUsersOK(s.T(), nil, nil, s.controller, identity.ID.String(), nil, nil)
	require.NotNil(s.T(), result)
	updatedContextInformation = result.Data.Attributes.ContextInformation

	// what was passed as non-nill should be intact.
	assert.Equal(s.T(), contextInformation["space"], updatedContextInformation["space"])

	// what was pass as nil should not be found!
	_, ok := updatedContextInformation["last_visited"]
	assert.Equal(s.T(), false, ok)
}

/*
	Pass no contextInformation and no one complains.
	This is as per general service behaviour.
*/

func (s *TestUsersSuite) TestUpdateUserOKWithoutContextInfo() {
	// given
	user := s.createRandomUser()
	identity := s.createRandomIdentity(user, account.KeycloakIDP)
	_, result := test.ShowUsersOK(s.T(), nil, nil, s.controller, identity.ID.String(), nil, nil)
	assert.Equal(s.T(), user.ID.String(), *result.Data.ID)
	assert.Equal(s.T(), user.FullName, *result.Data.Attributes.FullName)
	assert.Equal(s.T(), user.ImageURL, *result.Data.Attributes.ImageURL)
	assert.Equal(s.T(), identity.ProviderType, *result.Data.Attributes.ProviderType)
	assert.Equal(s.T(), identity.Username, *result.Data.Attributes.Username)
	// when
	newEmail := "TestUpdateUserOKWithoutContextInfo-" + uuid.NewV4().String() + "@email.com"
	newFullName := "TestUpdateUserOKWithoutContextInfo"
	newImageURL := "http://new.image.io/imageurl"
	newBio := "new bio"
	newProfileURL := "http://new.profile.url/url"
	secureService, secureController := s.SecuredController(identity)

	updateUsersPayload := createUpdateUsersPayloadWithoutContextInformation(&newEmail, &newFullName, &newBio, &newImageURL, &newProfileURL)
	test.UpdateUsersOK(s.T(), secureService.Context, secureService, secureController, updateUsersPayload)
}

func (s *TestUsersSuite) TestUpdateUserUnauthorized() {
	// given
	user := s.createRandomUser()
	identity := s.createRandomIdentity(user, account.KeycloakIDP)
	_, result := test.ShowUsersOK(s.T(), nil, nil, s.controller, identity.ID.String(), nil, nil)
	assert.Equal(s.T(), user.ID.String(), *result.Data.ID)
	assert.Equal(s.T(), user.FullName, *result.Data.Attributes.FullName)
	assert.Equal(s.T(), user.ImageURL, *result.Data.Attributes.ImageURL)
	assert.Equal(s.T(), identity.ProviderType, *result.Data.Attributes.ProviderType)
	assert.Equal(s.T(), identity.Username, *result.Data.Attributes.Username)
	newEmail := "TestUpdateUserUnauthorized-" + uuid.NewV4().String() + "@email.com"
	newFullName := "TestUpdateUserUnauthorized"
	newImageURL := "http://new.image.io/imageurl"
	newBio := "new bio"
	newProfileURL := "http://new.profile.url/url"
	contextInformation := map[string]interface{}{
		"last_visited": "yesterday",
		"space":        "3d6dab8d-f204-42e8-ab29-cdb1c93130ad",
	}
	//secureController, secureService := createSecureController(t, identity)
	updateUsersPayload := createUpdateUsersPayload(&newEmail, &newFullName, &newBio, &newImageURL, &newProfileURL, contextInformation)
	// when/then
	test.UpdateUsersUnauthorized(s.T(), context.Background(), nil, s.controller, updateUsersPayload)
}

func (s *TestUsersSuite) TestShowUserOK() {
	// given user
	user := s.createRandomUser()
	identity := s.createRandomIdentity(user, account.KeycloakIDP)
	// when
	res, result := test.ShowUsersOK(s.T(), nil, nil, s.controller, identity.ID.String(), nil, nil)
	// then
	assertUser(s.T(), result.Data, user, identity)
	assertSingleUserResponseHeaders(s.T(), res, result, user)
}

func (s *TestUsersSuite) TestShowUserOKUsingExpiredIfModifedSinceHeader() {
	// given user
	user := s.createRandomUser()
	identity := s.createRandomIdentity(user, account.KeycloakIDP)
	// when
	ifModifiedSince := user.UpdatedAt.Add(-1 * time.Hour)
	res, result := test.ShowUsersOK(s.T(), nil, nil, s.controller, identity.ID.String(), &ifModifiedSince, nil)
	// then
	assertUser(s.T(), result.Data, user, identity)
	assertSingleUserResponseHeaders(s.T(), res, result, user)
}

func (s *TestUsersSuite) TestShowUserOKUsingExpiredIfNoneMatchHeader() {
	// given user
	user := s.createRandomUser()
	identity := s.createRandomIdentity(user, account.KeycloakIDP)
	// when
	ifNoneMatch := "foo"
	res, result := test.ShowUsersOK(s.T(), nil, nil, s.controller, identity.ID.String(), nil, &ifNoneMatch)
	// then
	assertUser(s.T(), result.Data, user, identity)
	assertSingleUserResponseHeaders(s.T(), res, result, user)
}

func (s *TestUsersSuite) TestShowUserNotModifiedUsingIfModifedSinceHeader() {
	// given user
	user := s.createRandomUser()
	identity := s.createRandomIdentity(user, account.KeycloakIDP)
	// when/then
	ifModifiedSince := user.UpdatedAt
	test.ShowUsersNotModified(s.T(), nil, nil, s.controller, identity.ID.String(), &ifModifiedSince, nil)
}

func (s *TestUsersSuite) TestShowUserNotModifiedUsingIfNoneMatchHeader() {
	// given user
	user := s.createRandomUser()
	identity := s.createRandomIdentity(user, account.KeycloakIDP)
	// when/then
	ifNoneMatch := app.GenerateEntityTag(user)
	test.ShowUsersNotModified(s.T(), nil, nil, s.controller, identity.ID.String(), nil, &ifNoneMatch)
}

func (s *TestUsersSuite) TestListUsersOK() {
	// given user1
	user1 := s.createRandomUser()
	identity11 := s.createRandomIdentity(user1, account.KeycloakIDP)
	s.createRandomIdentity(user1, "github-test")
	// given user2
	user2 := s.createRandomUser()
	identity2 := s.createRandomIdentity(user2, account.KeycloakIDP)
	// when
	res, result := test.ListUsersOK(s.T(), nil, nil, s.controller, nil, nil)
	// then
	assertUser(s.T(), findUser(user1.ID, result.Data), user1, identity11)
	assertUser(s.T(), findUser(user2.ID, result.Data), user2, identity2)
	assertMultiUsersResponseHeaders(s.T(), res, user2)
}

func (s *TestUsersSuite) TestListUsersOKUsingExpiredIfModifiedSinceHeader() {
	// given user1
	user1 := s.createRandomUser()
	identity11 := s.createRandomIdentity(user1, account.KeycloakIDP)
	s.createRandomIdentity(user1, "github-test")
	// given user2
	user2 := s.createRandomUser()
	identity2 := s.createRandomIdentity(user2, account.KeycloakIDP)
	// when
	ifModifiedSinceHeader := user2.UpdatedAt.Add(-1 * time.Hour)
	res, result := test.ListUsersOK(s.T(), nil, nil, s.controller, &ifModifiedSinceHeader, nil)
	// then
	assertUser(s.T(), findUser(user1.ID, result.Data), user1, identity11)
	assertUser(s.T(), findUser(user2.ID, result.Data), user2, identity2)
	assertMultiUsersResponseHeaders(s.T(), res, user2)
}

func (s *TestUsersSuite) TestListUsersOKUsingExpiredIfNoneMatchHeader() {
	// given user1
	user1 := s.createRandomUser()
	identity11 := s.createRandomIdentity(user1, account.KeycloakIDP)
	s.createRandomIdentity(user1, "github-test")
	// given user2
	user2 := s.createRandomUser()
	identity2 := s.createRandomIdentity(user2, account.KeycloakIDP)
	// when
	ifNoneMatch := "foo"
	res, result := test.ListUsersOK(s.T(), nil, nil, s.controller, nil, &ifNoneMatch)
	// then
	assertUser(s.T(), findUser(user1.ID, result.Data), user1, identity11)
	assertUser(s.T(), findUser(user2.ID, result.Data), user2, identity2)
	assertMultiUsersResponseHeaders(s.T(), res, user2)
}
func (s *TestUsersSuite) TestListUsersNotModifiedUsingIfModifiedSinceHeader() {
	// given user1
	user1 := s.createRandomUser()
	s.createRandomIdentity(user1, account.KeycloakIDP)
	s.createRandomIdentity(user1, "github-test")
	// given user2
	user2 := s.createRandomUser()
	s.createRandomIdentity(user2, account.KeycloakIDP)
	// when/then
	ifModifiedSinceHeader := user2.UpdatedAt
	test.ListUsersNotModified(s.T(), nil, nil, s.controller, &ifModifiedSinceHeader, nil)
}

func (s *TestUsersSuite) TestListUsersNotModifiedUsingIfNoneMatchHeader() {
	// given user1
	user1 := s.createRandomUser()
	s.createRandomIdentity(user1, account.KeycloakIDP)
	s.createRandomIdentity(user1, "github-test")
	// given user2
	user2 := s.createRandomUser()
	s.createRandomIdentity(user2, account.KeycloakIDP)
	_, allUsers := test.ListUsersOK(s.T(), nil, nil, s.controller, nil, nil)
	// when/then
	ifNoneMatch := s.generateUsersTag(*allUsers)
	test.ListUsersNotModified(s.T(), nil, nil, s.controller, nil, &ifNoneMatch)
}

func (s *TestUsersSuite) createRandomUser() account.User {
	user := account.User{
		Email:    uuid.NewV4().String() + "primaryForUpdat7e@example.com",
		FullName: "A test user",
		ImageURL: "someURLForUpdate",
		ID:       uuid.NewV4(),
	}
	err := s.userRepo.Create(context.Background(), &user)
	require.Nil(s.T(), err)
	return user
}

func (s *TestUsersSuite) createRandomIdentity(user account.User, providerType string) account.Identity {
	profile := "foobarforupdate.com/" + uuid.NewV4().String() + "/" + user.ID.String()
	identity := account.Identity{
		Username:     "TestUpdateUserIntegration123" + uuid.NewV4().String(),
		ProviderType: providerType,
		ProfileURL:   &profile,
		User:         user,
		UserID:       account.NullUUID{UUID: user.ID, Valid: true},
	}
	err := s.identityRepo.Create(context.Background(), &identity)
	require.Nil(s.T(), err)
	return identity
}

func findUser(id uuid.UUID, userData []*app.UserData) *app.UserData {
	for _, user := range userData {
		if *user.ID == id.String() {
			return user
		}
	}
	return nil
}

func assertUser(t *testing.T, actual *app.UserData, expectedUser account.User, expectedIdentity account.Identity) {
	require.NotNil(t, actual)
	assert.Equal(t, expectedUser.ID.String(), *actual.ID)
	assert.Equal(t, expectedIdentity.Username, *actual.Attributes.Username)
	assert.Equal(t, expectedIdentity.ProviderType, *actual.Attributes.ProviderType)
	assert.Equal(t, expectedUser.FullName, *actual.Attributes.FullName)
	assert.Equal(t, expectedUser.ImageURL, *actual.Attributes.ImageURL)
	assert.Equal(t, expectedUser.Email, *actual.Attributes.Email)
}

func assertSingleUserResponseHeaders(t *testing.T, res http.ResponseWriter, appUser *app.User, modelUser account.User) {
	require.NotNil(t, res.Header()[app.LastModified])
	assert.Equal(t, getUserUpdatedAt(*appUser).String(), res.Header()[app.LastModified][0])
	require.NotNil(t, res.Header()[app.CacheControl])
	assert.Equal(t, app.MaxAge+"=300", res.Header()[app.CacheControl][0])
	require.NotNil(t, res.Header()[app.ETag])
	assert.Equal(t, app.GenerateEntityTag(modelUser), res.Header()[app.ETag][0])
}

func assertMultiUsersResponseHeaders(t *testing.T, res http.ResponseWriter, lastCreatedUser account.User) {
	require.NotNil(t, res.Header()[app.LastModified])
	assert.Equal(t, lastCreatedUser.UpdatedAt.Truncate(time.Second).UTC().String(), res.Header()[app.LastModified][0])
	require.NotNil(t, res.Header()[app.CacheControl])
	assert.Equal(t, app.MaxAge+"=300", res.Header()[app.CacheControl][0])
	require.NotNil(t, res.Header()[app.ETag])
}

func createUpdateUsersPayload(email, fullName, bio, imageURL, profileURL *string, contextInformation map[string]interface{}) *app.UpdateUsersPayload {
	return &app.UpdateUsersPayload{
		Data: &app.UpdateUserData{
			Type: "identities",
			Attributes: &app.UserDataAttributes{
				Email:              email,
				FullName:           fullName,
				Bio:                bio,
				ImageURL:           imageURL,
				URL:                profileURL,
				ContextInformation: contextInformation,
			},
		},
	}
}

func createUpdateUsersPayloadWithoutContextInformation(email, fullName, bio, imageURL, profileURL *string) *app.UpdateUsersPayload {
	return &app.UpdateUsersPayload{
		Data: &app.UpdateUserData{
			Type: "identities",
			Attributes: &app.UserDataAttributes{
				Email:    email,
				FullName: fullName,
				Bio:      bio,
				ImageURL: imageURL,
				URL:      profileURL,
			},
		},
	}
}

func getUserUpdatedAt(appUser app.User) time.Time {
	return appUser.Data.Attributes.UpdatedAt.Truncate(time.Second).UTC()
}

func (s *TestUsersSuite) generateUsersTag(allUsers app.UserArray) string {
	entities := make([]app.ConditionalResponseEntity, len(allUsers.Data))
	for i, user := range allUsers.Data {
		userID, err := uuid.FromString(*user.ID)
		require.Nil(s.T(), err)
		entities[i] = account.User{
			ID: userID,
			Lifecycle: gormsupport.Lifecycle{
				UpdatedAt: *user.Attributes.UpdatedAt,
			},
		}
	}
	logrus.Info("Users: ", len(allUsers.Data), " -> ETag: ", app.GenerateEntitiesTag(entities))
	return app.GenerateEntitiesTag(entities)
}
