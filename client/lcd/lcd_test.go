package lcd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wrsp "github.com/tepleton/wrsp/types"
	cryptoKeys "github.com/tepleton/go-crypto/keys"
	"github.com/tepleton/tepleton/p2p"
	ctypes "github.com/tepleton/tepleton/rpc/core/types"
	dbm "github.com/tepleton/tmlibs/db"
	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/tepleton-sdk/baseapp"
	"github.com/tepleton/tepleton-sdk/client"
	keys "github.com/tepleton/tepleton-sdk/client/keys"
	"github.com/tepleton/tepleton-sdk/examples/basecoin/app"
	"github.com/tepleton/tepleton-sdk/server"
)

func TestKeys(t *testing.T) {
	prepareClient(t)

	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// empty keys
	req, err := http.NewRequest("GET", "/keys", nil)
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
	body := res.Body.String()
	require.Equal(t, body, "[]", "Expected an empty array")

	// add key
	addr := createKey(t, r)
	assert.Len(t, addr, 40, "Returned address has wrong format", res.Body.String())

	// existing keys
	req, err = http.NewRequest("GET", "/keys", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var m [1]keys.KeyOutput
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)

	assert.Equal(t, m[0].Name, "test", "Did not serve keys name correctly")
	assert.Equal(t, m[0].Address, addr, "Did not serve keys Address correctly")

	// select key
	req, _ = http.NewRequest("GET", "/keys/test", nil)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var m2 keys.KeyOutput
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&m2)

	assert.Equal(t, m2.Name, "test", "Did not serve keys name correctly")
	assert.Equal(t, m2.Address, addr, "Did not serve keys Address correctly")

	// update key
	var jsonStr = []byte(`{"old_password":"1234567890", "new_password":"12345678901"}`)
	req, err = http.NewRequest("PUT", "/keys/test", bytes.NewBuffer(jsonStr))
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// here it should say unauthorized as we changed the password before
	req, err = http.NewRequest("PUT", "/keys/test", bytes.NewBuffer(jsonStr))
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code, res.Body.String())

	// delete key
	jsonStr = []byte(`{"password":"12345678901"}`)
	req, err = http.NewRequest("DELETE", "/keys/test", bytes.NewBuffer(jsonStr))
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
}

func TestVersion(t *testing.T) {
	prepareClient(t)
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// node info
	req, err := http.NewRequest("GET", "/version", nil)
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// TODO fix regexp
	// reg, err := regexp.Compile(`v\d+\.\d+\.\d+(-dev)?`)
	// require.Nil(t, err)
	// match := reg.MatchString(res.Body.String())
	// assert.True(t, match, res.Body.String())
	assert.Equal(t, "0.11.1-dev", res.Body.String())
}

func TestNodeStatus(t *testing.T) {
	ch := server.StartServer(t)
	defer close(ch)
	prepareClient(t)

	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// node info
	req, err := http.NewRequest("GET", "/node_info", nil)
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m p2p.NodeInfo
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse node info")

	assert.NotEqual(t, p2p.NodeInfo{}, m, "res: %v", res)

	// syncing
	req, err = http.NewRequest("GET", "/syncing", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.Equal(t, "true", res.Body.String())
}

func TestBlock(t *testing.T) {
	ch := server.StartServer(t)
	defer close(ch)
	prepareClient(t)

	cdc := app.MakeCodec()
	r := initRouter(cdc)

	req, err := http.NewRequest("GET", "/blocks/latest", nil)
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m ctypes.ResultBlock
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultBlock{}, m)

	req, err = http.NewRequest("GET", "/blocks/1", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.NotEqual(t, ctypes.ResultBlock{}, m)

	req, err = http.NewRequest("GET", "/blocks/2", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestValidators(t *testing.T) {
	ch := server.StartServer(t)
	defer close(ch)

	prepareClient(t)
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	req, err := http.NewRequest("GET", "/validatorsets/latest", nil)
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m ctypes.ResultValidators
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultValidators{}, m)

	req, err = http.NewRequest("GET", "/validatorsets/1", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.NotEqual(t, ctypes.ResultValidators{}, m)

	req, err = http.NewRequest("GET", "/validatorsets/2", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

//__________________________________________________________
// helpers

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

func prepareClient(t *testing.T) {
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(t.Name(), defaultLogger(), db)
	viper.Set(client.FlagNode, "localhost:46657")

	header := wrsp.Header{Height: 1}
	app.BeginBlock(wrsp.RequestBeginBlock{Header: header})
	app.Commit()
}

// setupViper creates a homedir to run inside,
// and returns a cleanup function to defer
func setupViper() func() {
	rootDir, err := ioutil.TempDir("", "mock-sdk-cmd")
	if err != nil {
		panic(err) // fuck it!
	}
	viper.Set("home", rootDir)
	return func() {
		os.RemoveAll(rootDir)
	}
}

func startServer(t *testing.T) {
	defer setupViper()()
	// init server
	initCmd := server.InitCmd(mock.GenInitOptions, log.NewNopLogger())
	err := initCmd.RunE(nil, nil)
	require.NoError(t, err)

	// start server
	viper.Set("with-tepleton", true)
	startCmd := server.StartCmd(mock.NewApp, log.NewNopLogger())
	timeout := time.Duration(3) * time.Second

	err = runOrTimeout(startCmd, timeout)
	require.NoError(t, err)
}

// copied from server/start_test.go
func runOrTimeout(cmd *cobra.Command, timeout time.Duration) error {
	done := make(chan error)
	go func(out chan<- error) {
		// this should NOT exit
		err := cmd.RunE(nil, nil)
		if err != nil {
			out <- err
		}
		out <- fmt.Errorf("start died for unknown reasons")
	}(done)
	timer := time.NewTimer(timeout)

	select {
	case err := <-done:
		return err
	case <-timer.C:
		return nil
	}
}

func createKey(t *testing.T, r http.Handler) string {
	var jsonStr = []byte(`{"name":"test", "password":"1234567890"}`)
	req, err := http.NewRequest("POST", "/keys", bytes.NewBuffer(jsonStr))
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

	addr := res.Body.String()
	return addr
}
