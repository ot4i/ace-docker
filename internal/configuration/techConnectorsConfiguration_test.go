package configuration

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ot4i/ace-docker/common/logger"
)

var processMqAccountRestore = processMqAccount
var processMqConnectorAccountsRestore = processMqConnectorAccounts
var processJdbcConnectorAccountsRestore = processJdbcConnectorAccounts
var mqAccount1Sha = "21659a449a678a235f7c14b8d8a196feedc0691124556624e57f2de60b5729d9"
var jdbcAccount1Sha = "569be438050b9875e183d092b133fa55e0ea57cfcdabe566c2d5613ab8492d50"

func TestSetupTechConnectorsConfigurations(t *testing.T) {

	var reset = func() {
		processMqAccount = func(log logger.LoggerInterface, basedir string, account MQAccountInfo, isDesignerAuthoringMode bool) error {
			panic("should be mocked")
		}

		processMqConnectorAccounts = func(log logger.LoggerInterface, basedir string, accounts []AccountInfo) error {
			panic("should be mocked")
		}

		processJdbcConnectorAccounts = func(log logger.LoggerInterface, basedir string, accounts []AccountInfo) error {
			panic("should be mocked")
		}
	}

	var restore = func() {
		processMqAccount = processMqAccountRestore
		processMqConnectorAccounts = processMqConnectorAccountsRestore
		processJdbcConnectorAccounts = processJdbcConnectorAccountsRestore
	}

	var accountsYaml = `
accounts:
  mq:
  - name: mq-account-1
    endpoint: {}
    credentials:
      authType: BASIC
      queueManager: QM1
      hostname: localhost1
      port: 123
      username: abc
      password: xyz
      channelName: CH.1
    applicationType: online
    applicationVersion: v1
    authType: BASIC
    default: true
`

	t.Run("when accounts.yml contains mq accounts, process all mq accounts", func(t *testing.T) {

		reset()
		defer restore()

		processMqConnectorAccounts = func(log logger.LoggerInterface, basedir string, accounts []AccountInfo) error {
			assert.NotNil(t, accounts)
			assert.Equal(t, 1, len(accounts))
			assert.Equal(t, "mq-account-1", accounts[0].Name)
			assert.NotNil(t, accounts[0].Credentials)

			return nil
		}

		err := SetupTechConnectorsConfigurations(testLogger, testBaseDir, []byte(accountsYaml))

		assert.Nil(t, err)
	})

	t.Run("when converting accounts.yml to json failed, returns nil ", func(t *testing.T) {

		var invalidYaml = `
accounts:
  mq:
  - name: test1
  invalid-entry: abc`

		reset()
		defer restore()

		err := SetupTechConnectorsConfigurations(testLogger, testBaseDir, []byte(invalidYaml))

		assert.Nil(t, err)
	})

	t.Run("when accounts yaml contains both jdbc and mq accounts, processes both mq and jdbc account", func(t *testing.T) {
		var accountsYaml = `
accounts:
  jdbc:
  - name: jdbc-account-a
    endpoint: {}
    credentials:
      dbType: 'IBM Db2 Linux, UNIX, or Windows (LUW) - client managed'
      dbName: testdb
      hostname: dbhost
      port: '50000'
      username: user1
      password: pwd1
    applicationType: online
    applicationVersion: v1
    authType: BASIC
    default: true
  mq:
  - name: mq-account-b
    endpoint: {}
    credentials:
      authType: BASIC
      queueManager: QM1
      hostname: localhost1
      port: '123'
      username: abc
      password: xyz
      channelName: CH.1
    applicationType: online
    applicationVersion: v1
    authType: BASIC
    default: true
`

		reset()
		defer restore()

		processJdbcConnectorAccounts = func(log logger.LoggerInterface, basedir string, accounts []AccountInfo) error {
			assert.NotNil(t, accounts)
			assert.Equal(t, 1, len(accounts))
			assert.Equal(t, "jdbc-account-a", accounts[0].Name)
			assert.NotNil(t, accounts[0].Credentials)

			return nil
		}

		processMqConnectorAccounts = func(log logger.LoggerInterface, basedir string, accounts []AccountInfo) error {
			assert.NotNil(t, accounts)
			assert.Equal(t, 1, len(accounts))
			assert.Equal(t, "mq-account-b", accounts[0].Name)
			assert.NotNil(t, accounts[0].Credentials)

			return nil
		}

		err := SetupTechConnectorsConfigurations(testLogger, testBaseDir, []byte(accountsYaml))

		assert.Nil(t, err)
	})
}

func TestProcessJdbcAccounts(t *testing.T) {
	var setDSNForJDBCApplicationRestore = setDSNForJDBCApplication
	var buildJDBCPoliciesRestore = buildJDBCPolicies
	var reset = func() {
		setDSNForJDBCApplication = func(log logger.LoggerInterface, basedir string, jdbcAccounts []JdbcAccountInfo) error {
			panic("should be mocked")
		}

		buildJDBCPolicies = func(log logger.LoggerInterface, basedir string, jdbcAccounts []JdbcAccountInfo) error {
			panic("should be mocked")
		}
	}

	var restore = func() {
		setDSNForJDBCApplication = setDSNForJDBCApplicationRestore
		buildJDBCPolicies = buildJDBCPoliciesRestore
	}

	accounts := []AccountInfo{
		{
			Name:        "acc 1",
			Credentials: json.RawMessage(`{"authType":"BASIC","dbType":"Oracle","hostname":"abc","port":"123","Username":"user1","Password":"password1","dbName":"test"}`)}}

	unmarshalledJDBCAccounts := []JdbcAccountInfo{{
		Name: "acc 1",
		Credentials: JdbcCredentials{
			AuthType: "BASIC",
			DbType:   "Oracle",
			Hostname: "abc",
			Port:     "123",
			Username: "user1",
			Password: "password1",
			DbName:   "test"}}}

	reset()
	defer restore()

	t.Run("Returns error when failed to setDSN for jdbc account", func(t *testing.T) {

		setDSNError := errors.New("set dsn failed")
		setDSNForJDBCApplication = func(log logger.LoggerInterface, basedir string, jdbcAccounts []JdbcAccountInfo) error {
			return setDSNError
		}

		errorReturned := processJdbcConnectorAccounts(testLogger, testBaseDir, accounts)

		assert.Equal(t, setDSNError, errorReturned)
	})

	t.Run("When setDSN succeeded, build jdbc policies", func(t *testing.T) {
		setDSNForJDBCApplication = func(log logger.LoggerInterface, basedir string, jdbcAccounts []JdbcAccountInfo) error {
			assert.Equal(t, unmarshalledJDBCAccounts, jdbcAccounts)

			return nil
		}

		t.Run("Returns error when buildJdbcPolicies failed", func(t *testing.T) {

			buildPolicyError := errors.New("set dsn failed")
			buildJDBCPolicies = func(log logger.LoggerInterface, basedir string, jdbcAccounts []JdbcAccountInfo) error {
				return buildPolicyError
			}

			errorReturned := processJdbcConnectorAccounts(testLogger, testBaseDir, accounts)

			assert.Equal(t, buildPolicyError, errorReturned)
		})

		t.Run("Returns nil when build jdbcPolicies succeeded", func(t *testing.T) {

			buildJDBCPolicies = func(log logger.LoggerInterface, basedir string, jdbcAccounts []JdbcAccountInfo) error {
				assert.Equal(t, unmarshalledJDBCAccounts, jdbcAccounts)
				return nil
			}

			errorReturned := processJdbcConnectorAccounts(testLogger, testBaseDir, accounts)

			assert.Nil(t, errorReturned)
		})
	})

}

func TestSetDSNForJDBCApplication(t *testing.T) {
	jdbAccounts := []JdbcAccountInfo{{
		Name: "acc 1",
		Credentials: JdbcCredentials{
			AuthType: "BASIC",
			DbType:   "Oracle",
			Hostname: "abc",
			Port:     "123",
			Username: "user1",
			Password: "password1",
			DbName:   "test"}}}

	var internalRunSetdbparmsCommandRestore = internalRunSetdbparmsCommand
	var reset = func() {
		internalRunSetdbparmsCommand = func(log logger.LoggerInterface, command string, params []string) error {
			panic("should be mocked")
		}
	}

	var restore = func() {
		internalRunSetdbparmsCommand = internalRunSetdbparmsCommandRestore
	}

	reset()
	defer restore()

	t.Run("invokes mqsisetdbparmas command with resource name which is sha of account", func(t *testing.T) {

		resourceName := "jdbc::" + jdbcAccount1Sha
		workdir := "'" + testBaseDir + string(os.PathSeparator) + workdirName + "'"
		internalRunSetdbparmsCommand = func(log logger.LoggerInterface, command string, params []string) error {
			assert.Equal(t, "mqsisetdbparms", command)
			assert.Equal(t, []string{"'-n'", resourceName, "'-u'", jdbAccounts[0].Credentials.Username.(string), "'-p'", jdbAccounts[0].Credentials.Password.(string), "'-w'", workdir}, params)
			return nil
		}

		setDSNForJDBCApplication(testLogger, testBaseDir, jdbAccounts)

	})

	t.Run("when mqsisetdbparmas command failed returns error", func(t *testing.T) {

		err := errors.New("run setdb params failed")
		internalRunSetdbparmsCommand = func(log logger.LoggerInterface, command string, params []string) error {
			return err
		}

		errReturned := setDSNForJDBCApplication(testLogger, testBaseDir, jdbAccounts)

		assert.Equal(t, err, errReturned)
	})

	t.Run("when mqsisetdbparmas command succeeds returns nil", func(t *testing.T) {

		internalRunSetdbparmsCommand = func(log logger.LoggerInterface, command string, params []string) error {
			return nil
		}

		errReturned := setDSNForJDBCApplication(testLogger, testBaseDir, jdbAccounts)

		assert.Nil(t, errReturned)
	})
}

func TestBuildJDBCPolicies(t *testing.T) {

	jdbAccounts := []JdbcAccountInfo{{
		Name: "acc 1",
		Credentials: JdbcCredentials{
			AuthType: "BASIC",
			DbType:   "Oracle",
			Hostname: "abc",
			Port:     "123",
			Username: "user1",
			Password: "password1",
			DbName:   "test"}}}

	var osStatRestore = osStat
	var osIsNotExistRestore = osIsNotExist
	var osMkdirAllRestore = osMkdirAll
	var transformXMLTemplateRestore = transformXMLTemplate
	var ioutilWriteFileRestore = ioutilWriteFile

	var reset = func() {
		osStat = func(dirName string) (os.FileInfo, error) {
			panic("should be mocked")
		}
		osIsNotExist = func(err error) bool {
			panic("should be mocked")
		}

		osMkdirAll = func(dirName string, filePerm os.FileMode) error {
			panic("should be mocked")
		}

		transformXMLTemplate = func(xmlTemplate string, context interface{}) (string, error) {
			panic("should be mocked")
		}

		ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
			panic("should be mocked")
		}
	}

	var restore = func() {
		osStat = osStatRestore
		osIsNotExist = osIsNotExistRestore
		osMkdirAll = osMkdirAllRestore
		transformXMLTemplate = transformXMLTemplateRestore
		ioutilWriteFile = ioutilWriteFileRestore
	}

	policyDirName := testBaseDir + string(os.PathSeparator) + workdirName + string(os.PathSeparator) + "overrides" + string(os.PathSeparator) + "gen.jdbcConnectorPolicies"

	reset()
	defer restore()

	t.Run("Returns error when failed to create 'gen.jdbcConnectorPolicies' directory if not exists", func(t *testing.T) {

		osStatError := errors.New("osStat failed")

		osStat = func(dirName string) (os.FileInfo, error) {
			assert.Equal(t, policyDirName, dirName)
			return nil, osStatError
		}

		osIsNotExist = func(err error) bool {
			assert.Equal(t, osStatError, err)
			return true
		}

		osMakeDirError := errors.New("os make dir error failed")
		osMkdirAll = func(dirPath string, perm os.FileMode) error {
			assert.Equal(t, policyDirName, dirPath)
			assert.Equal(t, os.ModePerm, perm)
			return osMakeDirError
		}

		errorReturned := buildJDBCPolicies(testLogger, testBaseDir, jdbAccounts)

		assert.Equal(t, osMakeDirError, errorReturned)

		t.Run("When creation of 'gen.jdbcConnectorPolicies' succeeds,  transforms policy xml template", func(t *testing.T) {

			osMkdirAll = func(dirPath string, perm os.FileMode) error {
				assert.Equal(t, policyDirName, dirPath)
				assert.Equal(t, os.ModePerm, perm)
				return nil
			}

			t.Run("Returns error when failed to transform policy xml template", func(t *testing.T) {

				transformError := errors.New("transform xml template failed")

				transformXMLTemplate = func(xmlTemplate string, context interface{}) (string, error) {
					return "", transformError
				}

				errorReturned := buildJDBCPolicies(testLogger, testBaseDir, jdbAccounts)
				assert.Equal(t, transformError, errorReturned)
			})

			t.Run("when transform xml template succeeds, writes transforrmed policy xml in 'gen.MQPolicies' dir with the file name of account name", func(t *testing.T) {

				trasformedXML := "abc"
				policyName := "oracle-" + jdbcAccount1Sha

				transformXMLTemplate = func(xmlTemplate string, context interface{}) (string, error) {

					assert.Equal(t, map[string]interface{}{
						"policyName":              policyName,
						"dbName":                  jdbAccounts[0].Credentials.DbName,
						"dbType":                  "oracle",
						"jdbcClassName":           "com.ibm.appconnect.jdbc.oracle.OracleDriver",
						"jdbcType4DataSourceName": "com.ibm.appconnect.jdbcx.oracle.OracleDataSource",
						"jdbcURL":                 "jdbc:ibmappconnect:oracle://abc:123;DatabaseName=test;user=[user];password=[password];loginTimeout=40;FetchDateAsTimestamp=false",
						"hostname":                jdbAccounts[0].Credentials.Hostname,
						"port":                    convertToNumber(jdbAccounts[0].Credentials.Port),
						"maxPoolSize":             convertToNumber(jdbAccounts[0].Credentials.MaxPoolSize),
						"securityIdentity":        jdbcAccount1Sha,
					}, context)

					return trasformedXML, nil
				}

				t.Run("Returns error when failed to write to account name policy file", func(t *testing.T) {

					writeFileError := errors.New("write file failed")
					policyFileNameForAccount1 := string(policyDirName + string(os.PathSeparator) + policyName + ".policyxml")
					ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
						assert.Equal(t, policyFileNameForAccount1, filename)
						assert.Equal(t, trasformedXML, string(data))
						assert.Equal(t, os.ModePerm, perm)
						return writeFileError
					}

					errorReturned := buildJDBCPolicies(testLogger, testBaseDir, jdbAccounts)
					assert.Equal(t, writeFileError, errorReturned)
				})

				t.Run("When write policyxml succeeds, writes policy descriptor", func(t *testing.T) {

					t.Run("Returns error when write policy descriptor failed", func(t *testing.T) {

						policyDescriptorWriteError := errors.New("policy descriptor write failed")
						ioutilWriteFile = func(filenameWithPath string, data []byte, perm os.FileMode) error {
							_, fileName := filepath.Split(filenameWithPath)
							if fileName == "policy.descriptor" {
								return policyDescriptorWriteError
							}

							return nil
						}

						errorReturned := buildJDBCPolicies(testLogger, testBaseDir, jdbAccounts)
						assert.Equal(t, policyDescriptorWriteError, errorReturned)
					})

					t.Run("Returns nil when write policy descriptor succeeeded", func(t *testing.T) {

						ioutilWriteFile = func(filenameWithPath string, data []byte, perm os.FileMode) error {
							return nil
						}

						errorReturned := buildJDBCPolicies(testLogger, testBaseDir, jdbAccounts)
						assert.Nil(t, errorReturned)
					})
				})

			})
		})
	})
}

func TestProcessMqConnectorAccounts(t *testing.T) {

	var reset = func() {
		processMqAccount = func(log logger.LoggerInterface, baseDir string, mqAccount MQAccountInfo, isDesignerAuthoringMode bool) error {
			panic("should be mocked")
		}
	}
	var restore = func() {
		processMqAccount = processMqAccountRestore
		os.Unsetenv("DEVELOPMENT_MODE")
	}

	t.Run("Process all mq accounts", func(t *testing.T) {

		reset()
		defer restore()

		mqAccounts := []AccountInfo{
			{Name: "acc 1", Credentials: json.RawMessage(`{"authType":"BASIC","Username":"abc","Password":"xyz"}`)},
			{Name: "acc 2", Credentials: json.RawMessage(`{"authType":"BASIC","Username":"u1","Password":"p1"}`)}}

		var callCount = 0
		processMqAccount = func(log logger.LoggerInterface, baseDir string, mqAccount MQAccountInfo, isDesignerAuthoringMode bool) error {

			if callCount == 0 {
				assert.Equal(t, "acc 1", mqAccount.Name)
				assert.Equal(t, "BASIC", mqAccount.Credentials.AuthType)
				assert.Equal(t, "abc", mqAccount.Credentials.Username)
				assert.Equal(t, "xyz", mqAccount.Credentials.Password)
			}

			if callCount == 1 {
				assert.Equal(t, "acc 2", mqAccount.Name)
				assert.Equal(t, "BASIC", mqAccount.Credentials.AuthType)
				assert.Equal(t, "u1", mqAccount.Credentials.Username)
				assert.Equal(t, "p1", mqAccount.Credentials.Password)
			}

			callCount++
			return nil
		}

		error := processMqConnectorAccounts(testLogger, testBaseDir, mqAccounts)

		assert.Equal(t, 2, callCount)
		assert.Nil(t, error)
	})

	t.Run("Returns error if any processMqAccount returned error", func(t *testing.T) {

		reset()
		defer restore()

		mqAccounts := []AccountInfo{
			{Name: "acc 1", Credentials: json.RawMessage(`{"authType":"BASIC","Username":"abc","Password":"xyz"}`)},
			{Name: "acc 2", Credentials: json.RawMessage(`{"authType":"BASIC","Username":"u1","Password":"p1"}`)}}

		var callCount = 0
		var err = errors.New("process mq account failed")
		processMqAccount = func(log logger.LoggerInterface, baseDir string, mqAccount MQAccountInfo, isDesignerAuthoringMode bool) error {
			if callCount == 1 {
				return err
			}

			callCount++
			return nil
		}

		errReturned := processMqConnectorAccounts(testLogger, testBaseDir, mqAccounts)

		assert.Equal(t, err, errReturned)
	})

	t.Run("When running in authoring mode, invokes processAllAccounts with isDesignerAuthoringMode flag true", func(t *testing.T) {
		reset()
		defer restore()

		os.Setenv("DEVELOPMENT_MODE", "true")

		mqAccounts := []AccountInfo{
			{Name: "acc 1", Credentials: json.RawMessage(`{"authType":"BASIC","Username":"abc","Password":"xyz"}`)}}

		processMqAccount = func(log logger.LoggerInterface, baseDir string, mqAccount MQAccountInfo, isDesignerAuthoringMode bool) error {
			assert.True(t, isDesignerAuthoringMode)
			return nil
		}

		err := processMqConnectorAccounts(testLogger, testBaseDir, mqAccounts)

		assert.Nil(t, err)
	})

	t.Run("When not running in authoring mode, invokes processAllAccounts with isDesignerAuthoringMode flag false", func(t *testing.T) {
		reset()
		defer restore()

		os.Unsetenv("DEVELOPMENT_MODE")

		mqAccounts := []AccountInfo{
			{Name: "acc 1", Credentials: json.RawMessage(`{"authType":"BASIC","Username":"abc","Password":"xyz"}`)}}

		processMqAccount = func(log logger.LoggerInterface, baseDir string, mqAccount MQAccountInfo, isDesignerAuthoringMode bool) error {
			assert.False(t, isDesignerAuthoringMode)
			return nil
		}

		err := processMqConnectorAccounts(testLogger, testBaseDir, mqAccounts)

		assert.Nil(t, err)
	})
}

func TestProcessMqAccount(t *testing.T) {

	var createMqAccountDbParamsRestore = createMqAccountDbParams
	var createMQPolicyRestore = createMQPolicy
	var createMQFlowBarOverridesPropertiesRestore = createMQFlowBarOverridesProperties

	var reset = func() {
		createMqAccountDbParams = func(log logger.LoggerInterface, basedir string, mqAccount MQAccountInfo) error {
			panic("should be mocked")
		}
		createMQPolicy = func(log logger.LoggerInterface, basedir string, mqAccount MQAccountInfo) error {
			panic("should be mocked")
		}
		createMQFlowBarOverridesProperties = func(log logger.LoggerInterface, basedir string, mqAccount MQAccountInfo) error {
			panic("should be mocked")
		}
		os.Setenv("ACE_CONTENT_SERVER_URL", "http://localhost/a.bar")
	}

	var restore = func() {
		createMqAccountDbParams = createMqAccountDbParamsRestore
		createMQPolicy = createMQPolicyRestore
		createMQFlowBarOverridesProperties = createMQFlowBarOverridesPropertiesRestore
		os.Unsetenv("ACE_CONTENT_SERVER_URL")
	}

	mqAccount := getMqAccount("acc-1")

	reset()
	defer restore()

	t.Run("when creates db params fails returns error", func(t *testing.T) {

		err := errors.New("create db params failed")
		createMqAccountDbParams = func(log logger.LoggerInterface, basedir string, mqAccountP MQAccountInfo) error {
			return err
		}

		errReturned := processMqAccount(testLogger, testBaseDir, mqAccount, false)

		assert.Equal(t, err, errReturned)
	})

	t.Run("when creates db params succeeds, creates policy", func(t *testing.T) {

		createMqAccountDbParams = func(log logger.LoggerInterface, basedir string, mqAccountP MQAccountInfo) error {
			assert.Equal(t, mqAccount, mqAccountP)
			return nil
		}

		t.Run("when running in designer authoring mode, doesn't deploy policy and doesn't create bar overrides file", func(t *testing.T) {

			isDesignerAuthoringMode := true
			createMQPolicy = func(log logger.LoggerInterface, basedir string, mqAccountP MQAccountInfo) error {
				assert.Fail(t, "Should not create policies in authoring mode")
				return nil
			}

			createMQFlowBarOverridesProperties = func(log logger.LoggerInterface, basedir string, mqAccountP MQAccountInfo) error {
				assert.Fail(t, "Should not create policies in authoring mode")
				return nil
			}

			err := processMqAccount(testLogger, testBaseDir, mqAccount, isDesignerAuthoringMode)

			assert.Nil(t, err)
		})

		t.Run("When mulitple bar files present, returns MQ connector not supported error", func(t *testing.T) {

			err := errors.New("IBM MQ Connector not supported for muliple bar files")

			os.Setenv("ACE_CONTENT_SERVER_URL", "http://localhost/a.bar,http://localhost/b.bar")

			errReturned := processMqAccount(testLogger, testBaseDir, mqAccount, false)

			assert.Equal(t, err, errReturned)
		})

		t.Run("When single bar present, create mq policy", func(t *testing.T) {
			os.Setenv("ACE_CONTENT_SERVER_URL", "http://localhost/a.bar")

			t.Run("when create policy fails, returns error", func(t *testing.T) {

				err := errors.New("create mq policy failed")
				createMQPolicy = func(log logger.LoggerInterface, basedir string, mqAccountP MQAccountInfo) error {
					return err
				}

				errReturned := processMqAccount(testLogger, testBaseDir, mqAccount, false)

				assert.Equal(t, err, errReturned)
			})

			t.Run("when create policy succeeds, creates bar overrides files", func(t *testing.T) {
				createMQPolicy = func(log logger.LoggerInterface, basedir string, mqAccountP MQAccountInfo) error {
					assert.Equal(t, mqAccount, mqAccountP)
					return nil
				}

				t.Run("when create bar ovrrides failed, returns error", func(t *testing.T) {
					err := errors.New("create bar overrides failed")
					createMQFlowBarOverridesProperties = func(log logger.LoggerInterface, basedir string, mqAccountP MQAccountInfo) error {
						return err
					}

					errReturned := processMqAccount(testLogger, testBaseDir, mqAccount, false)

					assert.Equal(t, err, errReturned)
				})

				t.Run("when create bar ovrrides succeeds, returns nil", func(t *testing.T) {

					createMQFlowBarOverridesProperties = func(log logger.LoggerInterface, basedir string, mqAccountP MQAccountInfo) error {
						assert.Equal(t, mqAccount, mqAccountP)
						return nil
					}

					err := processMqAccount(testLogger, testBaseDir, mqAccount, false)

					assert.Nil(t, err)
				})
			})
		})
	})
}

func TestCreateMqAccountDbParams(t *testing.T) {

	var internalRunSetdbparmsCommandRestore = internalRunSetdbparmsCommand
	var reset = func() {
		internalRunSetdbparmsCommand = func(log logger.LoggerInterface, command string, params []string) error {
			panic("should be mocked")
		}
	}

	var restore = func() {
		internalRunSetdbparmsCommand = internalRunSetdbparmsCommandRestore
	}

	t.Run("invokes mqsisetdbparmas command with resource name which is sha of account, username and password arguments", func(t *testing.T) {

		reset()
		defer restore()

		mqAccount := getMqAccount("acc1")
		resourceName := "mq::gen_" + mqAccount1Sha
		workdir := "'" + testBaseDir + string(os.PathSeparator) + workdirName + "'"
		internalRunSetdbparmsCommand = func(log logger.LoggerInterface, command string, params []string) error {
			assert.Equal(t, "mqsisetdbparms", command)
			assert.Equal(t, []string{"'-n'", resourceName, "'-u'", "'" + mqAccount.Credentials.Username + "'", "'-p'", "'" + mqAccount.Credentials.Password + "'", "'-w'", workdir}, params)
			return nil
		}

		createMqAccountDbParams(testLogger, testBaseDir, mqAccount)
	})

	t.Run("when mqsisetdbparmas command failed returns error", func(t *testing.T) {

		reset()
		defer restore()

		mqAccount := getMqAccount("acc1")
		err := errors.New("run setdb params failed")
		internalRunSetdbparmsCommand = func(log logger.LoggerInterface, command string, params []string) error {
			return err
		}

		errReturned := createMqAccountDbParams(testLogger, testBaseDir, mqAccount)

		assert.Equal(t, err, errReturned)
	})

	t.Run("when mqsisetdbparmas command succeeds returns nil", func(t *testing.T) {

		reset()
		defer restore()

		mqAccount := getMqAccount("acc1")
		internalRunSetdbparmsCommand = func(log logger.LoggerInterface, command string, params []string) error {
			return nil
		}

		errReturned := createMqAccountDbParams(testLogger, testBaseDir, mqAccount)

		assert.Nil(t, errReturned)
	})

	t.Run("when username and password fields are empty doesn't execute setdbparams command", func(t *testing.T) {

		reset()
		defer restore()

		internalRunSetdbparmsCommand = func(log logger.LoggerInterface, command string, params []string) error {
			assert.Fail(t, "setdbparams should not be invoked when username and password fields are empty")
			return nil
		}

		mqAccount := getMqAccount("acc1")
		mqAccount.Credentials.Username = ""
		mqAccount.Credentials.Password = ""

		errReturned := createMqAccountDbParams(testLogger, testBaseDir, mqAccount)

		assert.Nil(t, errReturned)
	})
}

func TestCreateMqPolicy(t *testing.T) {

	var osStatRestore = osStat
	var osIsNotExistRestore = osIsNotExist
	var osMkdirAllRestore = osMkdirAll
	var transformXMLTemplateRestore = transformXMLTemplate
	var ioutilWriteFileRestore = ioutilWriteFile
	mqAccount := getMqAccount("acc-1")

	var reset = func() {
		osStat = func(dirName string) (os.FileInfo, error) {
			panic("should be mocked")
		}
		osIsNotExist = func(err error) bool {
			panic("should be mocked")
		}

		osMkdirAll = func(dirName string, filePerm os.FileMode) error {
			panic("should be mocked")
		}

		transformXMLTemplate = func(xmlTemplate string, context interface{}) (string, error) {
			panic("should be mocked")
		}

		ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
			panic("should be mocked")
		}
	}

	var restore = func() {
		osStat = osStatRestore
		osIsNotExist = osIsNotExistRestore
		osMkdirAll = osMkdirAllRestore
		transformXMLTemplate = transformXMLTemplateRestore
		ioutilWriteFile = ioutilWriteFileRestore
	}

	policyDirName := testBaseDir + string(os.PathSeparator) + workdirName + string(os.PathSeparator) + "overrides" + string(os.PathSeparator) + "gen.MQPolicies"

	t.Run("Returns error when failed to create 'gen.MQPolicies' directory if not exists", func(t *testing.T) {

		reset()
		defer restore()

		osStatError := errors.New("osStat failed")

		osStat = func(dirName string) (os.FileInfo, error) {
			assert.Equal(t, policyDirName, dirName, mqAccount)
			return nil, osStatError
		}

		osIsNotExist = func(err error) bool {
			assert.Equal(t, osStatError, err)
			return true
		}

		osMakeDirError := errors.New("os make dir error failed")
		osMkdirAll = func(dirPath string, perm os.FileMode) error {
			assert.Equal(t, policyDirName, dirPath)
			assert.Equal(t, os.ModePerm, perm)
			return osMakeDirError
		}

		errorReturned := createMQPolicy(testLogger, testBaseDir, mqAccount)

		assert.Equal(t, osMakeDirError, errorReturned)

		t.Run("When creation of 'gen.MQPolicies' succeeds,  transforms xml template", func(t *testing.T) {

			osMkdirAll = func(dirPath string, perm os.FileMode) error {
				assert.Equal(t, policyDirName, dirPath)
				assert.Equal(t, os.ModePerm, perm)
				return nil
			}

			t.Run("Returns error when failed to transform policy xml template", func(t *testing.T) {

				transformError := errors.New("transform xml template failed")
				policyName := "account_1"

				transformXMLTemplate = func(xmlTemplate string, context interface{}) (string, error) {
					assert.Equal(t, map[string]interface{}{
						"policyName":       policyName,
						"queueManager":     mqAccount.Credentials.QueueManager,
						"hostName":         mqAccount.Credentials.Hostname,
						"port":             mqAccount.Credentials.Port,
						"channelName":      mqAccount.Credentials.ChannelName,
						"securityIdentity": "gen_" + mqAccount1Sha,
					}, context)

					return "", transformError
				}

				errorReturned := createMQPolicy(testLogger, testBaseDir, mqAccount)
				assert.Equal(t, transformError, errorReturned)
			})

			t.Run("Sets security idenity empty with mq account username and password are empty", func(t *testing.T) {

				transformError := errors.New("transform xml template failed")
				policyName := "account_1"

				transformXMLTemplate = func(xmlTemplate string, context interface{}) (string, error) {
					assert.Equal(t, map[string]interface{}{
						"policyName":       policyName,
						"queueManager":     mqAccount.Credentials.QueueManager,
						"hostName":         mqAccount.Credentials.Hostname,
						"port":             mqAccount.Credentials.Port,
						"channelName":      mqAccount.Credentials.ChannelName,
						"securityIdentity": "",
					}, context)

					return "", transformError
				}

				mqAccountWithEmptyCredentials := getMqAccountWithEmptyCredentials()

				errorReturned := createMQPolicy(testLogger, testBaseDir, mqAccountWithEmptyCredentials)
				assert.Equal(t, transformError, errorReturned)
			})

			t.Run("when transform xml template succeeds, writes transforrmed policy xml in 'gen.MQPolicies' dir with the file name of account name", func(t *testing.T) {

				trasformedXML := "abc"
				transformXMLTemplate = func(xmlTemplate string, context interface{}) (string, error) {
					return trasformedXML, nil
				}

				t.Run("Returns error when failed to write to account name policy file", func(t *testing.T) {

					writeFileError := errors.New("write file failed")
					policyFileNameForAccount1 := string(policyDirName + string(os.PathSeparator) + "account_1.policyxml")
					ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
						assert.Equal(t, policyFileNameForAccount1, filename)
						assert.Equal(t, trasformedXML, string(data))
						assert.Equal(t, os.ModePerm, perm)
						return writeFileError
					}

					errorReturned := createMQPolicy(testLogger, testBaseDir, mqAccount)
					assert.Equal(t, writeFileError, errorReturned)
				})

				t.Run("When write policyxml succeeds, writes policy descriptor", func(t *testing.T) {

					t.Run("Returns error when write policy descriptor failed", func(t *testing.T) {

						policyDescriptorWriteError := errors.New("policy descriptor write failed")
						ioutilWriteFile = func(filenameWithPath string, data []byte, perm os.FileMode) error {
							_, fileName := filepath.Split(filenameWithPath)
							if fileName == "policy.descriptor" {
								return policyDescriptorWriteError
							}

							return nil
						}

						errorReturned := createMQPolicy(testLogger, testBaseDir, mqAccount)
						assert.Equal(t, policyDescriptorWriteError, errorReturned)
					})

					t.Run("Returns nil when write policy descriptor succeeeded", func(t *testing.T) {

						ioutilWriteFile = func(filenameWithPath string, data []byte, perm os.FileMode) error {
							return nil
						}

						errorReturned := createMQPolicy(testLogger, testBaseDir, mqAccount)
						assert.Nil(t, errorReturned)
					})
				})

			})
		})
	})
}

func TestCreateMQFlowBarOverridesPropertiest(t *testing.T) {

	var osStatRestore = osStat
	var osIsNotExistRestore = osIsNotExist
	var osMkdirAllRestore = osMkdirAll
	var internalAppendFileRestore = internalAppendFile
	mqAccount := getMqAccount("acc-1")

	var reset = func() {
		osStat = func(dirName string) (os.FileInfo, error) {
			panic("should be mocked")
		}
		osIsNotExist = func(err error) bool {
			panic("should be mocked")
		}

		osMkdirAll = func(dirName string, filePerm os.FileMode) error {
			panic("should be mocked")
		}

		internalAppendFile = func(filename string, data []byte, perm os.FileMode) error {
			panic("should be mocked")
		}
	}

	var restore = func() {
		osStat = osStatRestore
		osIsNotExist = osIsNotExistRestore
		osMkdirAll = osMkdirAllRestore
		internalAppendFile = internalAppendFileRestore
	}

	reset()
	defer restore()

	t.Run("Returns error when failed to create bar_overrides dir if not exist", func(t *testing.T) {

		barOverridesDir := "/home/aceuser/initial-config/bar_overrides"
		osStatError := errors.New("osStat failed")

		osStat = func(dirName string) (os.FileInfo, error) {
			assert.Equal(t, barOverridesDir, dirName)
			return nil, osStatError
		}

		osIsNotExist = func(err error) bool {
			assert.Equal(t, osStatError, err)
			return true
		}

		osMakeDirError := errors.New("os make dir error failed")

		osMkdirAll = func(dirPath string, perm os.FileMode) error {
			assert.Equal(t, barOverridesDir, dirPath)
			assert.Equal(t, os.ModePerm, perm)
			return osMakeDirError
		}

		errorReturned := createMQFlowBarOverridesProperties(testLogger, testBaseDir, mqAccount)

		assert.Equal(t, osMakeDirError, errorReturned)

		t.Run("when create barfile overrides dir succeeded, creates barfile.properties with httpendpoint of mq runtime flow", func(t *testing.T) {

			osMkdirAll = func(dirPath string, perm os.FileMode) error {
				return nil
			}

			t.Run("Returns error when failed to write barfile.properties", func(t *testing.T) {

				barPropertiesFile := "/home/aceuser/initial-config/bar_overrides/barfile.properties"
				appendFileError := errors.New("append file failed")

				flowHTTPEndPoint := "gen.mq_account_1#HTTPInput.URLSpecifier=/_lcp-mq-connect_" + mqAccount1Sha + "\n"
				internalAppendFile = func(fileName string, fileContent []byte, filePerm os.FileMode) error {
					assert.Equal(t, barPropertiesFile, fileName)
					assert.Equal(t, flowHTTPEndPoint, string(fileContent))
					return appendFileError
				}

				errorReturned := createMQFlowBarOverridesProperties(testLogger, testBaseDir, mqAccount)
				assert.Equal(t, appendFileError, errorReturned)
			})

			t.Run("Returns nil when write barfile.properties succeeded", func(t *testing.T) {

				internalAppendFile = func(fileName string, fileContent []byte, filePerm os.FileMode) error {
					return nil
				}

				errorReturned := createMQFlowBarOverridesProperties(testLogger, testBaseDir, mqAccount)
				assert.Nil(t, errorReturned)
			})
		})
	})
}

func getMqAccount(accountName string) MQAccountInfo {
	return MQAccountInfo{
		Name: "account-1",
		Credentials: MQCredentials{
			AuthType:     "BASIC",
			QueueManager: "QM1",
			Hostname:     "host1",
			Port:         123,
			Username:     "abc",
			Password:     "xyz",
			ChannelName:  "testchannel"}}
}

func getMqAccountWithEmptyCredentials() MQAccountInfo {
	return MQAccountInfo{
		Name: "account-1",
		Credentials: MQCredentials{
			AuthType:     "BASIC",
			QueueManager: "QM1",
			Hostname:     "host1",
			Port:         123,
			Username:     "",
			Password:     "",
			ChannelName:  "testchannel"}}
}
