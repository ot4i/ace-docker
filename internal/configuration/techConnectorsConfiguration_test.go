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
var runMqakmCommandRestore = runMqakmCommand
var runOpenSslCommandRestore = runOpenSslCommand
var createMqAccountsKdbFileRestore = createMqAccountsKdbFile

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

		setupMqAccountsKdbFile = func(log logger.LoggerInterface) error {
			panic("should be mocked")
		}

		runOpenSslCommand = func(log logger.LoggerInterface, params []string) error {
			panic("should be mocked")
		}
	}

	var restore = func() {
		processMqAccount = processMqAccountRestore
		processMqConnectorAccounts = processMqConnectorAccountsRestore
		processJdbcConnectorAccounts = processJdbcConnectorAccountsRestore
		setupMqAccountsKdbFile = setupMqAccountsKdbFileRestore
		runMqakmCommand = runMqakmCommandRestore
		runOpenSslCommand = runOpenSslCommandRestore
		setupMqAccountsKdbFile = setupMqAccountsKdbFileRestore
		createMqAccountsKdbFile = createMqAccountsKdbFileRestore
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

	t.Run("Setups mq kdb file before processing accounts", func(t *testing.T) {

		reset()
		defer restore()

		t.Run("Returns error if setupMqaccounts failed", func(t *testing.T) {

			setupMqAccountsError := errors.New("set up mq accounts kdb failed")
			setupMqAccountsKdbFile = func(log logger.LoggerInterface) error {
				return setupMqAccountsError
			}

			err := SetupTechConnectorsConfigurations(testLogger, testBaseDir, []byte(accountsYaml))

			assert.Error(t, err)
		})

		t.Run("when setupMqaccounts succeeded,process accounts ", func(t *testing.T) {

			setupMqAccountsKdbFile = func(log logger.LoggerInterface) error {
				return nil
			}

			t.Run("when accounts.yml contains mq accounts, process all mq accounts", func(t *testing.T) {

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
		})
	})
}

func TestSetMqAccountsKdbFileTests(t *testing.T) {

	var reset = func() {
		ioutilReadFile = func(file string) ([]byte, error) {
			panic("should be mocked")
		}

		createMqAccountsKdbFile = func(log logger.LoggerInterface) error {
			panic("should be mocked")
		}

		ioutilWriteFile = func(file string, content []byte, perm os.FileMode) error {
			panic("should be mocked")
		}
	}

	var restore = func() {
		ioutilWriteFile = ioutilWriteFileRestore
		ioutilReadFile = ioutilReadFileRestore
		createMqAccountsKdbFile = createMqAccountsKdbFileRestore
	}

	reset()
	defer restore()

	t.Run("when no server.conf.yaml not found, continues to create kdb file", func(t *testing.T) {

		serverConfFilePath := "/home/aceuser/ace-server/overrides/server.conf.yaml"
		ioutilReadFile = func(file string) ([]byte, error) {
			assert.Equal(t, serverConfFilePath, file)
			return nil, errors.New("File not found")
		}

		t.Run("Returns error if failed to create kdb file", func(t *testing.T) {

			createMqAccountsKdbFile = func(log logger.LoggerInterface) error {
				return errors.New("create mq accounts kdb failed")
			}

			err := setupMqAccountsKdbFile(testLogger)

			assert.Error(t, err, "create mq accounts kdb failed")
		})
	})

	t.Run("when no server.conf.yaml found, continues to create kdb", func(t *testing.T) {

		t.Run("does not update server.conf if it already has mq key repository", func(t *testing.T) {
			serverConfContent := "BrokerRegistry:\n  mqKeyRepository: /home/aceuser/somekdbfile"

			ioutilReadFile = func(file string) ([]byte, error) {
				return []byte(serverConfContent), nil
			}

			ioutilWriteFile = func(file string, content []byte, perm os.FileMode) error {
				assert.Fail(t, "shouldn't update server conf ")
				return nil
			}

			setupMqAccountsKdbFile(testLogger)
		})

		t.Run("Returns error if failed to create kdb file when server conf doesnot contain mqKeyRepository", func(t *testing.T) {

			serverConfContent := ""

			ioutilReadFile = func(file string) ([]byte, error) {
				return []byte(serverConfContent), nil
			}

			createMqAccountsKdbFile = func(log logger.LoggerInterface) error {
				return errors.New("Failed to create kdb")
			}

			err := setupMqAccountsKdbFile(testLogger)

			assert.Error(t, err, "Failed to create kdb")
		})
	})

	t.Run("when create kdb succeeded, writes server conf yaml with mqKeyRepsitory", func(t *testing.T) {

		createMqAccountsKdbFile = func(log logger.LoggerInterface) error {
			return nil
		}

		t.Run("Return error when write to server.conf failed", func(t *testing.T) {

			ioutilWriteFile = func(file string, contents []byte, fileMode os.FileMode) error {
				return errors.New("server.conf write failed")
			}

			err := setupMqAccountsKdbFile(testLogger)

			assert.Error(t, err, "server.conf write failed")
		})

		t.Run("Return nil when write server.conf failed succeeded", func(t *testing.T) {

			ioutilWriteFile = func(file string, contents []byte, fileMode os.FileMode) error {
				serverConfContent := "BrokerRegistry:\n  mqKeyRepository: /home/aceuser/kdb/mq\n"
				serverConfFilePath := "/home/aceuser/ace-server/overrides/server.conf.yaml"

				assert.Equal(t, serverConfFilePath, file)
				assert.Equal(t, serverConfContent, string(contents))
				assert.Equal(t, os.FileMode(0644), fileMode)
				return nil
			}

			err := setupMqAccountsKdbFile(testLogger)

			assert.Nil(t, err)
		})
	})
}

func TestCreateMqAccountsKdbFileImpl(t *testing.T) {

	var reset = func() {
		osMkdirAll = func(path string, filePerm os.FileMode) error {
			panic("should be mocked")
		}
		runMqakmCommand = func(logger logger.LoggerInterface, cmdArgs []string) error {
			panic("should be mocked")
		}
	}

	var restore = func() {
		osMkdirAll = osMkdirAllRestore
		runMqakmCommand = runMqakmCommandRestore
	}

	reset()
	defer restore()

	t.Run("creats directory mq-connector for kdb file", func(t *testing.T) {

		t.Run("Returns error if failed to create dir", func(t *testing.T) {

			osMkdirAll = func(path string, filePerm os.FileMode) error {
				return errors.New("mkdir failed")
			}

			err := createMqAccountsKdbFile(testLogger)

			assert.Error(t, err, "Failed to create directory to create kdb file, Error mkdir failed")

		})

		t.Run("when create dir succeeded, creates kdb file ", func(t *testing.T) {

			osMkdirAll = func(path string, filePerm os.FileMode) error {
				dirPath := "/home/aceuser/kdb"

				assert.Equal(t, dirPath, path)
				assert.Equal(t, os.ModePerm, filePerm)

				return nil
			}

			t.Run("Returns error when runmqakam failed to create kdb", func(t *testing.T) {
				runMqakmCommand = func(logger logger.LoggerInterface, cmdArgs []string) error {

					assert.Equal(t, "-keydb", cmdArgs[0])
					assert.Equal(t, "-create", cmdArgs[1])
					assert.Equal(t, "-type", cmdArgs[2])
					assert.Equal(t, "cms", cmdArgs[3])
					assert.Equal(t, "-db", cmdArgs[4])
					assert.Equal(t, "/home/aceuser/kdb/mq.kdb", cmdArgs[5])
					assert.Equal(t, "-pw", cmdArgs[6])
					assert.NotEmpty(t, cmdArgs[7])

					return errors.New("runmqakam failed")
				}

				err := createMqAccountsKdbFile(testLogger)

				assert.Error(t, err, "Create kdb failed, Error runmqakam failed")
			})

			t.Run("when create kdb succeeded, creates stash file", func(t *testing.T) {

				t.Run("Returns error when create stash failed failed", func(t *testing.T) {

					runMqakmCommand = func(logger logger.LoggerInterface, cmdArgs []string) error {
						if cmdArgs[1] == "-stashpw" {
							assert.Equal(t, "-keydb", cmdArgs[0])
							assert.Equal(t, "-stashpw", cmdArgs[1])
							assert.Equal(t, "-type", cmdArgs[2])
							assert.Equal(t, "cms", cmdArgs[3])
							assert.Equal(t, "-db", cmdArgs[4])
							assert.Equal(t, "/home/aceuser/kdb/mq.kdb", cmdArgs[5])
							assert.Equal(t, "-pw", cmdArgs[6])
							assert.NotEmpty(t, cmdArgs[7])

							return errors.New("create sth file failed")
						}

						return nil
					}

					err := createMqAccountsKdbFile(testLogger)

					assert.Error(t, err, "Create sth failed, Error runmqakam failed")
				})

				t.Run("Returns nil when creation of stash file succeeded", func(t *testing.T) {

					runMqakmCommand = func(logger logger.LoggerInterface, cmdArgs []string) error {
						return nil
					}

					err := createMqAccountsKdbFile(testLogger)

					assert.Nil(t, err)
				})
			})
		})
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
	var importMqAccountCertificatesRestore = importMqAccountCertificates

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
		importMqAccountCertificates = func(log logger.LoggerInterface, mqAccount MQAccountInfo) error {
			panic("should be mocked")
		}
		os.Setenv("ACE_CONTENT_SERVER_URL", "http://localhost/a.bar")
	}

	var restore = func() {
		createMqAccountDbParams = createMqAccountDbParamsRestore
		createMQPolicy = createMQPolicyRestore
		createMQFlowBarOverridesProperties = createMQFlowBarOverridesPropertiesRestore
		importMqAccountCertificates = importMqAccountCertificatesRestore

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

	t.Run("when creates db params succeeds,import certificates", func(t *testing.T) {

		createMqAccountDbParams = func(log logger.LoggerInterface, basedir string, mqAccountP MQAccountInfo) error {
			return nil
		}

		t.Run("Returns error when import certificates failed", func(t *testing.T) {

			importError := errors.New("import certificates failed")
			importMqAccountCertificates = func(log logger.LoggerInterface, mqAccount MQAccountInfo) error {
				return importError
			}

			errReturned := processMqAccount(testLogger, testBaseDir, mqAccount, false)
			assert.Equal(t, importError, errReturned)
		})

		t.Run("when import certificates succeeded, creates policies", func(t *testing.T) {

			importMqAccountCertificates = func(log logger.LoggerInterface, mqAccount MQAccountInfo) error {
				return nil
			}

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
						"policyName":          policyName,
						"queueManager":        mqAccount.Credentials.QueueManager,
						"hostName":            mqAccount.Credentials.Hostname,
						"port":                mqAccount.Credentials.Port,
						"channelName":         mqAccount.Credentials.ChannelName,
						"securityIdentity":    "gen_" + mqAccount1Sha,
						"useSSL":              false,
						"sslPeerName":         "",
						"sslCipherSpec":       "",
						"sslCertificateLabel": "",
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
						"policyName":          policyName,
						"queueManager":        mqAccount.Credentials.QueueManager,
						"hostName":            mqAccount.Credentials.Hostname,
						"port":                mqAccount.Credentials.Port,
						"channelName":         mqAccount.Credentials.ChannelName,
						"securityIdentity":    "",
						"useSSL":              false,
						"sslPeerName":         "",
						"sslCipherSpec":       "",
						"sslCertificateLabel": "",
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

	t.Run("Returns error when failed to create workdir_overrides dir if not exist", func(t *testing.T) {

		barOverridesDir := "/home/aceuser/initial-config/workdir_overrides"
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

				barPropertiesFile := "/home/aceuser/initial-config/workdir_overrides/mqconnectorbarfile.properties"
				appendFileError := errors.New("append file failed")

				udfProperty := "gen.mq_account_1#mqRuntimeAccountId=account_1~" + mqAccount1Sha + "\n"
				internalAppendFile = func(fileName string, fileContent []byte, filePerm os.FileMode) error {
					assert.Equal(t, barPropertiesFile, fileName)
					assert.Equal(t, udfProperty, string(fileContent))
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

func TestImportMqAccountCertificatesTests(t *testing.T) {

	var convertMqAccountSingleLinePEMRestore = convertMqAccountSingleLinePEM
	var kdbFilePath = "/home/aceuser/kdb/mq.kdb"

	var reset = func() {
		osMkdirAll = func(dirName string, filePerm os.FileMode) error {
			panic("should be mocked")
		}
		runOpenSslCommand = func(log logger.LoggerInterface, cmdArgs []string) error {
			panic("should be mocked")
		}

		runMqakmCommand = func(logger logger.LoggerInterface, cmdArgs []string) error {
			panic("should be mocked")
		}

		convertMqAccountSingleLinePEM = func(pem string) (string, error) {
			panic("should be mocked")
		}
	}

	var restore = func() {
		osMkdirAll = osMkdirAllRestore
		runOpenSslCommand = runOpenSslCommandRestore
		runMqakmCommand = runMqakmCommandRestore
		convertMqAccountSingleLinePEM = convertMqAccountSingleLinePEMRestore
	}

	reset()
	defer restore()

	t.Run("Creates temp dir to import certificates", func(t *testing.T) {

		mqAccountWithSsl := getMqAccountWithSsl("acc1")

		t.Run("Returns error if failed to create temp dir", func(t *testing.T) {
			osMkdirAll = func(dirName string, filePerm os.FileMode) error {
				assert.Equal(t, "/tmp/mqssl-work", dirName)
				assert.Equal(t, os.ModePerm, filePerm)
				return errors.New("Create temp dir failed")
			}

			err := importMqAccountCertificates(testLogger, mqAccountWithSsl)

			assert.Error(t, err, "Failed to create mp dir to import certificates,Error: Create temp dir failed")

		})

		t.Run("when create temp dir succeeded", func(t *testing.T) {

			osMkdirAll = func(dirName string, filePerm os.FileMode) error {
				return nil
			}

			t.Run("imports server cerificates if present", func(t *testing.T) {

				t.Run("Returns error if the server certificate is not valid a PEM ", func(t *testing.T) {
					pemError := errors.New("Invalid pem error")
					convertMqAccountSingleLinePEM = func(content string) (string, error) {
						assert.Equal(t, mqAccountWithSsl.Credentials.ServerCertificate, content)
						return "", pemError
					}

					err := importMqAccountCertificates(testLogger, mqAccountWithSsl)

					assert.Equal(t, pemError, err)
				})

				t.Run("when server certificate is a valid PEM, writes converted PEM to a temp file to import", func(t *testing.T) {

					convertedPem := "abc"
					tempServerCertificatePemFile := "/tmp/mqssl-work/servercrt.pem"
					convertMqAccountSingleLinePEM = func(content string) (string, error) {
						return convertedPem, nil
					}

					t.Run("Returns error if write to temp file failed", func(t *testing.T) {

						writeError := errors.New("Wrie temp file failed")

						ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
							assert.Equal(t, tempServerCertificatePemFile, filename)
							assert.Equal(t, convertedPem, string(data))
							assert.Equal(t, os.ModePerm, perm)

							return writeError
						}

						err := importMqAccountCertificates(testLogger, mqAccountWithSsl)

						assert.Equal(t, writeError, err)
					})

					t.Run("when write to temp file succeeded, runs mqakam command to import certificate to kdb", func(t *testing.T) {
						ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
							return nil
						}

						t.Run("Return error if runmqakm command failed", func(t *testing.T) {
							cmdError := errors.New("command error")
							runMqakmCommand = func(log logger.LoggerInterface, cmdArgs []string) error {
								return cmdError
							}

							err := importMqAccountCertificates(testLogger, mqAccountWithSsl)

							assert.Equal(t, cmdError, err)

						})

						t.Run("Returns nil When runMqakm command succeeded to import certificate", func(t *testing.T) {

							runMqakmCommand = func(log logger.LoggerInterface, cmdArgs []string) error {
								assert.Equal(t, "-cert", cmdArgs[0])
								assert.Equal(t, "-add", cmdArgs[1])
								assert.Equal(t, "-db", cmdArgs[2])
								assert.Equal(t, kdbFilePath, cmdArgs[3])
								assert.Equal(t, "-stashed", cmdArgs[4])
								assert.Equal(t, "-file", cmdArgs[5])
								assert.Equal(t, tempServerCertificatePemFile, cmdArgs[6])
								return nil
							}

							err := importMqAccountCertificates(testLogger, mqAccountWithSsl)

							assert.Nil(t, err)
						})
					})

				})
			})

			t.Run("imports client cerificate if present", func(t *testing.T) {

				mqAccountWithMutualSsl := getMqAccountWithMutualAuthSsl("acc1")
				mqAccountWithMutualSsl.Credentials.ServerCertificate = ""

				t.Run("Returns error if the client certificate label is empty", func(t *testing.T) {

					mqAccountWithEmptyCertLabel := getMqAccountWithMutualAuthSsl("acc1")
					mqAccountWithEmptyCertLabel.Credentials.ClientCertificateLabel = ""
					err := importMqAccountCertificates(testLogger, mqAccountWithEmptyCertLabel)

					assert.Error(t, err, "Certificate label should not be empty for acc1")
				})

				t.Run("Returns error if the client certificate is not valid a PEM ", func(t *testing.T) {
					pemError := errors.New("Invalid pem error")
					convertMqAccountSingleLinePEM = func(content string) (string, error) {
						assert.Equal(t, mqAccountWithMutualSsl.Credentials.ClientCertificate, content)
						return "", pemError
					}

					err := importMqAccountCertificates(testLogger, mqAccountWithMutualSsl)

					assert.Equal(t, pemError, err)
				})

				t.Run("when client certificate is a valid PEM, writes converted PEM to a temp file to import", func(t *testing.T) {

					convertedPem := "abc"
					tempClientCertificatePemFile := "/tmp/mqssl-work/clientcrt.pem"
					convertMqAccountSingleLinePEM = func(content string) (string, error) {
						return convertedPem, nil
					}

					t.Run("Returns error when write to temp file failed", func(t *testing.T) {

						writeError := errors.New("Write client pem temp file failed")

						ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
							assert.Equal(t, tempClientCertificatePemFile, filename)
							assert.Equal(t, convertedPem, string(data))
							assert.Equal(t, os.ModePerm, perm)

							return writeError
						}

						err := importMqAccountCertificates(testLogger, mqAccountWithMutualSsl)

						assert.Error(t, err, "Write client pem temp file failed")
					})

					t.Run("when write to temp file succeeded, creates pkcs12 file usingg openssl command to import to kdb", func(t *testing.T) {
						ioutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
							return nil
						}

						tempP12File := "/tmp/mqssl-work/clientcrt.p12"

						t.Run("runs openssl command to create p12 with required arguments when no password present", func(t *testing.T) {

							runOpenSslCommand = func(log logger.LoggerInterface, cmdArgs []string) error {
								assert.Equal(t, "pkcs12", cmdArgs[0])
								assert.Equal(t, "-export", cmdArgs[1])
								assert.Equal(t, "-out", cmdArgs[2])
								assert.Equal(t, tempP12File, cmdArgs[3])
								assert.Equal(t, "-passout", cmdArgs[4])
								assert.NotEmpty(t, cmdArgs[5])
								assert.NotEmpty(t, "-in", cmdArgs[6])
								assert.Equal(t, tempClientCertificatePemFile, cmdArgs[7])

								return errors.New("open ssl failed")
							}

							err := importMqAccountCertificates(testLogger, mqAccountWithMutualSsl)

							assert.NotNil(t, err)
						})

						t.Run("when password is parent runs openssl command password in flag", func(t *testing.T) {

							mqAccountWithPassword := getMqAccountWithMutualAuthSsl("acc1")
							mqAccountWithPassword.Credentials.ClientCertificatePassword = "p123"

							runOpenSslCommand = func(log logger.LoggerInterface, cmdArgs []string) error {

								assert.NotEmpty(t, "-passin", cmdArgs[8])
								assert.Equal(t, "pass:p123", cmdArgs[9])

								return errors.New("open ssl failed")
							}

							err := importMqAccountCertificates(testLogger, mqAccountWithPassword)

							assert.NotNil(t, err)
						})

						t.Run("Return error if openssl command failed to create p12 file failed", func(t *testing.T) {
							cmdError := errors.New("command error")
							runOpenSslCommand = func(log logger.LoggerInterface, cmdArgs []string) error {
								return cmdError
							}

							err := importMqAccountCertificates(testLogger, mqAccountWithMutualSsl)

							assert.Equal(t, cmdError, err)

						})

						t.Run("When creation of p12 file from pem succeeded runs runmqakm command to import p12 file to kdb", func(t *testing.T) {

							runOpenSslCommand = func(log logger.LoggerInterface, cmdArgs []string) error {
								return nil
							}

							t.Run("Returns error if runmqakm command failed to import p12 file", func(t *testing.T) {

								cmdErr := errors.New("runmqakm failed")
								runMqakmCommand = func(log logger.LoggerInterface, cmdArgs []string) error {
									return cmdErr
								}
								err := importMqAccountCertificates(testLogger, mqAccountWithMutualSsl)
								assert.Equal(t, cmdErr, err)
							})

							t.Run("Returns nil when runmqakm command succeeded", func(t *testing.T) {

								runMqakmCommand = func(log logger.LoggerInterface, cmdArgs []string) error {
									assert.Equal(t, "-cert", cmdArgs[0])
									assert.Equal(t, "-import", cmdArgs[1])
									assert.Equal(t, "-type", cmdArgs[2])
									assert.Equal(t, "p12", cmdArgs[3])
									assert.Equal(t, "-file", cmdArgs[4])
									assert.Equal(t, tempP12File, cmdArgs[5])
									assert.Equal(t, "-pw", cmdArgs[6])
									assert.NotEmpty(t, cmdArgs[7])
									assert.NotEmpty(t, "-target_type", cmdArgs[8])
									assert.NotEmpty(t, "cms", cmdArgs[9])
									assert.NotEmpty(t, "-target", cmdArgs[10])
									assert.Equal(t, kdbFilePath, cmdArgs[11])
									assert.NotEmpty(t, "-target_stashed", cmdArgs[12])

									return nil
								}

								err := importMqAccountCertificates(testLogger, mqAccountWithMutualSsl)
								assert.Nil(t, err)
							})
						})
					})
				})
			})
		})
	})
}

func TestConvertMQAccountSingleLinePEM(t *testing.T) {

	t.Run("Returns error if no PEM headers and footers present in the PEM line", func(t *testing.T) {
		pemLine := "abc no pem headers and footers"

		_, err := convertMqAccountSingleLinePEM(pemLine)

		assert.NotNil(t, err)
	})

	t.Run("Returns error if not a valid PEM", func(t *testing.T) {
		pemLineInvalidFooter := "-----BEGIN CERTIFICATE----- line1 line2 -----END CERTIFICATE"

		_, err := convertMqAccountSingleLinePEM(pemLineInvalidFooter)

		assert.NotNil(t, err)
	})

	t.Run("Returns error if not a valid PEM", func(t *testing.T) {
		pemLineInvalidHeader := "-----BEGIN----- line1 line2 -----END CERTIFICATE-----"

		_, err := convertMqAccountSingleLinePEM(pemLineInvalidHeader)

		assert.NotNil(t, err)
	})

	t.Run("Returns multiline PEM", func(t *testing.T) {
		expectedPem := "-----BEGIN CERTIFICATE-----\nline1\nline2\n-----END CERTIFICATE-----\n-----BEGIN ENCRYPTED PRIVATE KEY-----\nkeyline1\nkeyline2\n-----END ENCRYPTED PRIVATE KEY-----\n"
		pemLineInvalidHeader := "-----BEGIN CERTIFICATE----- line1 line2 -----END CERTIFICATE----- -----BEGIN ENCRYPTED PRIVATE KEY----- keyline1 keyline2 -----END ENCRYPTED PRIVATE KEY-----"

		multiLinePem, err := convertMqAccountSingleLinePEM(pemLineInvalidHeader)

		assert.Nil(t, err)
		assert.Equal(t, expectedPem, multiLinePem)

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

func getMqAccountWithSsl(accountName string) MQAccountInfo {
	return MQAccountInfo{
		Name: "account-1",
		Credentials: MQCredentials{
			AuthType:          "BASIC",
			QueueManager:      "QM1",
			Hostname:          "host1",
			Port:              123,
			Username:          "abc",
			Password:          "xyz",
			ChannelName:       "testchannel",
			ServerCertificate: "-----BEGIN CERTIFICATE----- Line1Content Line2Content -----END CERTIFICATE-----",
		}}
}

func getMqAccountWithMutualAuthSsl(accountName string) MQAccountInfo {
	return MQAccountInfo{
		Name: "account-1",
		Credentials: MQCredentials{
			AuthType:                  "BASIC",
			QueueManager:              "QM1",
			Hostname:                  "host1",
			Port:                      123,
			Username:                  "abc",
			Password:                  "xyz",
			ChannelName:               "testchannel",
			ServerCertificate:         "-----BEGIN CERTIFICATE----- Line1Content Line2Content -----END CERTIFICATE-----",
			ClientCertificate:         "-----BEGIN CERTIFICATE----- Line1Content Line2Content -----END CERTIFICATE-----",
			ClientCertificatePassword: "abcd",
			ClientCertificateLabel:    "somelabel",
		}}
}
