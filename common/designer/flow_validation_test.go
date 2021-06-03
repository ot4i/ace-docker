/*
© Copyright IBM Corporation 2020

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package designer

import (
	"os"
	"errors"
	"io"
	"io/ioutil"

	"github.com/ot4i/ace-docker/common/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var testLogger, _ = logger.NewLogger(os.Stdout, true, true, "test")

func TestGetConnectorLicenseToggleName(t *testing.T) {
	require.Equal(t, "connector-test", getConnectorLicenseToggleName("test"))
}

func TestFindDisabledConnectorInFlow(t *testing.T) {
	oldIsLicenseToggleEnabled := isLicenseToggleEnabled 
	isLicenseToggleEnabled = func(toggle string) bool {
		if toggle == "connector-test" {
			return true
		} else {
			return false
		}
	}
 	t.Run("When there are no unsupported connectors", func(t *testing.T) {
		testFlowDocument :=  flowDocument {
			integration {
				map[string]flowInterface {
					"trigger-interface1": {
						ConnectorType: "test",
					},
				},
				map[string]flowInterface {
					"action-interface1": {
						ConnectorType: "test",
					},
				},
			},
		}
		require.Equal(t, "", findDisabledConnectorInFlow(testFlowDocument, testLogger))
	})

	t.Run("When there are unsupported connectors in the trigger interface", func(t *testing.T) {
		testFlowDocument :=  flowDocument {
			integration {
				map[string]flowInterface {
					"trigger-interface1": {
						ConnectorType: "test",
					},
					"trigger-interface2": {
						ConnectorType: "foo",
					},
				},
				map[string]flowInterface {
					"action-interface2": {
						ConnectorType: "test",
					},
				},
			},
		}
		disabledConnectors := findDisabledConnectorInFlow(testFlowDocument, testLogger)
		require.Equal(t, "foo", disabledConnectors)
	})

	t.Run("When there are unsupported connectors in the action interface", func(t *testing.T) {
		testFlowDocument :=  flowDocument {
			integration {
				map[string]flowInterface {
					"trigger-interface1": {
						ConnectorType: "test",
					},
				},
				map[string]flowInterface {
					"action-interface1": {
						ConnectorType: "test",
					},
					"action-interface2": {
						ConnectorType: "bar",
					},
				},
			},
		}
		disabledConnectors := findDisabledConnectorInFlow(testFlowDocument, testLogger)
		require.Equal(t, "bar", disabledConnectors)
	})

	t.Run("When there are unsupported connectors in both the trigger interface and the action interface", func(t *testing.T) {
		testFlowDocument :=  flowDocument {
			integration {
				map[string]flowInterface {
					"trigger-interface1": {
						ConnectorType: "test",
					},
					"trigger-interface2": {
						ConnectorType: "foo",
					},
				},
				map[string]flowInterface {
					"action-interface1": {
						ConnectorType: "test",
					},
					"action-interface2": {
						ConnectorType: "bar",
					},
				},
			},
		}
		disabledConnectors := findDisabledConnectorInFlow(testFlowDocument, testLogger)
		require.Equal(t, "foo, bar", disabledConnectors)
	})

	isLicenseToggleEnabled = oldIsLicenseToggleEnabled
}

func TestCopy(t *testing.T) {
	basedir, err := os.Getwd()
	require.NoError(t, err)

	src := basedir + string(os.PathSeparator) + "src.txt"
	dst := basedir + string(os.PathSeparator) + "dst.txt"

	t.Run("When failing to open the src file", func (t *testing.T) {
		oldOsOpen := osOpen
		osOpen = func (string) (*os.File, error) {
			return nil, errors.New("Test")
		}
		err = copy(src, dst, testLogger)
		require.Error(t, err)
		osOpen = oldOsOpen
	})

	t.Run("When succeeding to open the src file", func (t *testing.T) {
		file, err := os.Create(src)
		assert.NoError(t, err)
		_, err = file.WriteString("src string")
		assert.NoError(t, err)

		t.Run("When failing to create the dst file", func (t *testing.T) {
			oldOsCreate := osCreate
			osCreate = func (string) (*os.File, error) {
				return nil, errors.New("Test")
			}

			err = copy(src, dst, testLogger)
			assert.Error(t, err)
			osCreate = oldOsCreate
		})

		t.Run("When succeeding to create the dst file", func (t *testing.T) {	
			t.Run("When the copying fails", func (t *testing.T) {
				oldIoCopy := ioCopy
				ioCopy = func (io.Writer, io.Reader) (int64, error) {
					return 0, errors.New("Test")
				}

				err = copy(src, dst, testLogger)
				assert.Error(t, err)
				ioCopy = oldIoCopy
			})

			t.Run("When the copying succeeds", func (t *testing.T) {
				err = copy(src, dst, testLogger)
				assert.NoError(t, err)

				contents, err := ioutil.ReadFile(dst)
				assert.NoError(t, err)
				assert.Equal(t, "src string", string(contents))
			})
		})
	})

	err = os.Remove(src)
	assert.NoError(t, err)
	err = os.Remove(dst)
	assert.NoError(t, err)
}

func TestReplaceFlow(t *testing.T) {
	basedir, err := os.Getwd()
	require.NoError(t, err)
	flow := "Test"	

	// setup folders
	err = os.Mkdir(basedir + string(os.PathSeparator) + "ace-server", 0777)
	assert.NoError(t, err)
	err = os.Mkdir(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run", 0777)
	assert.NoError(t, err)

	t.Run("When deleting the /run/<flow>PolicyProject folder fails", func (t *testing.T) {
		oldRemoveAll := removeAll
		removeAll = func (string) error {
			return errors.New("Fail delete /run/<flow>PolicyProject")
		}

		err := replaceFlow(flow, testLogger, basedir)
		require.Error(t, err)
		
		removeAll = oldRemoveAll
	})

	t.Run("When deleting the /run/<flow>PolicyProject folder does not fail", func (t *testing.T) {
		err = os.Mkdir(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + "PolicyProject", 0777)
		assert.NoError(t, err)

		t.Run("When renaming the /run/<flow> folder fails", func (t *testing.T) {
			oldRename := rename
			rename = func (string, string) error {
				return errors.New("Fail rename /run/<flow>")
			}

			err := replaceFlow(flow, testLogger, basedir)
			require.Error(t, err)
			
			rename = oldRename
		})

		t.Run("When renaming the /run/<flow> folder succeeds", func (t *testing.T) {
			t.Run("When deleting the /run/<flow>/gen folder fails", func (t *testing.T) {
				oldRemoveAll := removeAll
				count := 0
				removeAll = func (string) error {
					if count == 0 {
						return nil
					}
					count ++;
					return errors.New("Fail delete /run/<flow>/gen")
				}

				err := replaceFlow(flow, testLogger, basedir)
				require.Error(t, err)

				removeAll = oldRemoveAll
			})

			t.Run("When deleting the /run/<flow>/gen folder succeeds", func (t *testing.T) {
				var setupFlow = func () {
					err = os.Mkdir(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow, 0777)
					assert.NoError(t, err)
					_, err = os.Create(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + string(os.PathSeparator) + "test.msgflow")
					assert.NoError(t, err)
					_, err = os.Create(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + string(os.PathSeparator) + "test.subflow")
					assert.NoError(t, err)
					_, err = os.Create(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + string(os.PathSeparator) + "restapi.descriptor")
					assert.NoError(t, err)
					err = os.Mkdir(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + string(os.PathSeparator) + "gen", 0777)
					assert.NoError(t, err)
					_, err = os.Create(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + string(os.PathSeparator) + "gen" + string(os.PathSeparator) + "valid.msgflow")
					assert.NoError(t, err)
				}
		
				var cleanupFlow = func () {
					err = os.RemoveAll(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + "_invalid_license")
					assert.NoError(t, err)
				}
				
				t.Run("When recreating the /run/<flow>/gen folder fails", func (t *testing.T) {
					oldMkdir:= mkdir
					mkdir = func (string, os.FileMode) error {
						return errors.New("Fail recreate /run/<flow>/gen")
					}

					setupFlow()
					err := replaceFlow(flow, testLogger, basedir)
					require.Error(t, err)
					cleanupFlow()

					mkdir = oldMkdir
				})
				
				t.Run("When recreating the /run/<flow>/gen folder succeeds", func (t *testing.T) {
					t.Run("When reading the /temp/gen folder fails", func (t *testing.T) {
						oldReadDir := readDir
						readDir = func(string) ([]os.FileInfo, error) {
							return nil, errors.New("Fail read  /temp/gen")
						}
						setupFlow()
						err := replaceFlow(flow, testLogger, basedir)
						require.Error(t, err)
						cleanupFlow()
						readDir = oldReadDir
					})

					t.Run("When reading the /temp/gen folder succeeds", func (t *testing.T) {
						err = os.Mkdir(basedir + string(os.PathSeparator) + "temp", 0777)
						assert.NoError(t, err)
						_, err = os.Create(basedir + string(os.PathSeparator) + "temp" + string(os.PathSeparator) + "application.descriptor")
						assert.NoError(t, err)
						_, err = os.Create(basedir + string(os.PathSeparator) + "temp" + string(os.PathSeparator) + "invalid.flow")
						assert.NoError(t, err)
						err = os.Mkdir(basedir + string(os.PathSeparator) + "temp" + string(os.PathSeparator) + "gen", 0777)
						assert.NoError(t, err)
						_, err = os.Create(basedir + string(os.PathSeparator) + "temp" + string(os.PathSeparator) + "gen" + string(os.PathSeparator) + "invalid.msgflow")
						assert.NoError(t, err)

						oldCopy := copy
						t.Run("When copying the files under /temp/gen to /run/<flow>/gen fails", func(t *testing.T) {
							setupFlow()
							copy = func (string, string, logger.LoggerInterface) error {
								return errors.New("Fail copy /temp/gen to /run/<flow>/gen")
							}

							err := replaceFlow(flow, testLogger, basedir)
							require.Error(t, err)

							cleanupFlow()
							copy = oldCopy
						})

						t.Run("When copying the files under /temp/gen to /run/<flow>/gen succeeds", func(t *testing.T) {
							t.Run("When reading the /<flow> folder fails", func (t *testing.T) {
								oldReadDir := readDir
								count := 0
								readDir = func(string) ([]os.FileInfo, error) {
									if count == 0 {
										return nil, nil
									}
									count++;
									return nil, errors.New("Fail read /<flow>")
								}
		
								err := replaceFlow(flow, testLogger, basedir)
								require.Error(t, err)

								readDir = oldReadDir
							})

							t.Run("When reading the /<flow> folder succeeds", func (t *testing.T) {
								t.Run("When removing .msgflow, .subflow and, restapi.descriptor from /run/<flow> fails", func(t *testing.T) {
									oldRemove := remove
									remove = func (string) error {
										return errors.New("Fail removing .msgflow, .subflow and, restapi.descriptor")
									}
						
									err := replaceFlow(flow, testLogger, basedir)
									require.Error(t, err)

									remove = oldRemove
								})
		
								t.Run("When removing .msgflow, .subflow and, restapi.descriptor from /run/<flow> succeeds", func(t *testing.T) {
									t.Run("When replacing restapi.descriptor with application.descriptor fails", func(t *testing.T) {
										copy = func (src string, dst string, logger logger.LoggerInterface) error {
											return errors.New("Fail copying .msgflow, .subflow and, restapi.descriptor")
										}
										setupFlow()

										err := replaceFlow(flow, testLogger, basedir)
										require.Error(t, err)

										cleanupFlow()
										copy = oldCopy
									})
		
									t.Run("When replacing restapi.descriptor with application.descriptor succeeds", func(t *testing.T) {
										setupFlow()
										assert.True(t, dirExists(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + string(os.PathSeparator) + "gen" + string(os.PathSeparator)  + "valid.msgflow"))
										assert.False(t, dirExists(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + string(os.PathSeparator) + "gen" + string(os.PathSeparator)  + "invalid.msgflow"))
						
										assert.True(t, dirExists(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + string(os.PathSeparator)  + "restapi.descriptor"))
										assert.False(t, dirExists(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + string(os.PathSeparator)  + "application.descriptor"))
				
										err := replaceFlow(flow, testLogger, basedir)
										require.NoError(t, err)
										
										assert.False(t, dirExists(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + "_invalid_license" + string(os.PathSeparator) + "gen" + string(os.PathSeparator)  + "valid.msgflow"))
										assert.True(t, dirExists(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + "_invalid_license" + string(os.PathSeparator) + "gen" + string(os.PathSeparator)  + "invalid.msgflow"))
						
										assert.False(t, dirExists(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + "_invalid_license" + string(os.PathSeparator)  + "restapi.descriptor"))
										assert.True(t, dirExists(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + "_invalid_license" + string(os.PathSeparator)  + "application.descriptor"))
		
										cleanupFlow()
									})
								})
							})
						})
						copy = oldCopy
					})
				})
			})

			// rename = oldRename
		})

		err = os.RemoveAll(basedir + string(os.PathSeparator) + "ace-server")
		assert.NoError(t, err)
		err = os.RemoveAll(basedir + string(os.PathSeparator) + "temp")
		assert.NoError(t, err)
	})
}

func TestCleanupInvalidBarResources(t *testing.T) {
	basedir, err := os.Getwd()
	require.NoError(t, err)

	t.Run("When the directory exists", func(t *testing.T) {
		// setup folder
		err = os.Mkdir(basedir + string(os.PathSeparator) + "temp", 0777)
		assert.NoError(t, err)
		_, err = os.Create(basedir + string(os.PathSeparator) + "temp" + string(os.PathSeparator) + "test.bar")
		assert.NoError(t, err)

		// test function
		assert.True(t, dirExists(basedir + string(os.PathSeparator) + "temp" ))
		err = cleanupInvalidBarResources(basedir)
		assert.NoError(t, err)
		assert.False(t, dirExists(basedir + string(os.PathSeparator) + "temp"))
	})

	t.Run("When the directory does not exist", func(t *testing.T) {
		assert.False(t, dirExists(basedir + string(os.PathSeparator) + "temp"))
		err := cleanupInvalidBarResources(basedir)
		assert.NoError(t, err)
		assert.False(t, dirExists(basedir + string(os.PathSeparator) + "temp"))
	})
}

func TestIsFlowValid(t *testing.T) {
	t.Run("When failing to parse flow", func (t *testing.T) {
		findDisabledConnectorInFlowCalls := 0
		findDisabledConnectorInFlow = func (flowDocument, logger.LoggerInterface) string {
			findDisabledConnectorInFlowCalls++;
			return ""
		}

		_, err := IsFlowValid(testLogger, "Test", []byte("foo"))
		assert.Error(t, err)

		assert.Equal(t, 0, findDisabledConnectorInFlowCalls)
	})

	t.Run("When findDisabledConnectorInFlow does not return a connector", func (t *testing.T) {
		findDisabledConnectorInFlowCalls := 0
		findDisabledConnectorInFlow = func (flowDocument, logger.LoggerInterface) string {
			findDisabledConnectorInFlowCalls++;
			return ""
		}

		valid, err := IsFlowValid(testLogger, "Test", []byte(flowDocumentYAML))
		assert.NoError(t, err)
		assert.True(t, valid)

		assert.Equal(t, 1, findDisabledConnectorInFlowCalls)
	})

	t.Run("When findDisabledConnectorInFlow does return a connector", func (t *testing.T) {
		findDisabledConnectorInFlowCalls := 0
		findDisabledConnectorInFlow = func (flowDocument, logger.LoggerInterface) string {
			findDisabledConnectorInFlowCalls++;
			return "test"
		}

		valid, err := IsFlowValid(testLogger, "Test", []byte(flowDocumentYAML))
		assert.NoError(t, err)
		assert.False(t, valid)

		assert.Equal(t, 1, findDisabledConnectorInFlowCalls)
	})
}

func TestValidateFlow(t *testing.T) {
	basedir, err := os.Getwd()
	require.NoError(t, err)

	t.Run("When the folders are not setup", func (t *testing.T) {
		err := ValidateFlows(testLogger, basedir)
		require.Error(t, err)
	})

	t.Run("When the folders are setup", func (t *testing.T) {
		// setup folders
		err = os.Mkdir(basedir + string(os.PathSeparator) + "ace-server", 0777)
		assert.NoError(t, err)
		err = os.Mkdir(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run", 0777)
		assert.NoError(t, err)
		// designer flow
		err = os.Mkdir(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + "Test", 0777)
		assert.NoError(t, err)
		err = os.Mkdir(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + "TestPolicyProject", 0777)
		assert.NoError(t, err)

		file, err := os.Create(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + "Test" +  string(os.PathSeparator) + "Test.yaml")
		assert.NoError(t, err)
		_, err = file.WriteString(flowDocumentYAML)
		assert.NoError(t, err)
		file.Close()
		
		oldIsFlowValid := IsFlowValid

		t.Run("When IsFlowValid returns an error", func (t *testing.T) {
			IsFlowValidCalls := 0
			IsFlowValid = func (logger.LoggerInterface, string, []byte) (bool, error) {
				IsFlowValidCalls++;
				return true, errors.New("IsFlowValid fails")
			}
	
			oldreplaceFlowNumCalls := 0
			oldreplaceFlow := replaceFlow
			replaceFlow = func (string, logger.LoggerInterface, string) error {
				oldreplaceFlowNumCalls++
				return nil
			}
			
			err := ValidateFlows(testLogger, basedir)
			assert.Error(t, err)

			assert.Equal(t, 1, IsFlowValidCalls)
			assert.Equal(t, 0, oldreplaceFlowNumCalls)

			replaceFlow = oldreplaceFlow
		})

		t.Run("When IsFlowValid returns true", func (t *testing.T) {
			// setup the toolkit flow every time, since it gets deleted at the end of the function
			err = os.Mkdir(basedir + string(os.PathSeparator) + "temp", 0777)
			assert.NoError(t, err)

			IsFlowValidCalls := 0
			IsFlowValid = func (logger.LoggerInterface, string, []byte) (bool, error) {
				IsFlowValidCalls++;
				return true, nil
			}
	
			oldreplaceFlowNumCalls := 0
			oldreplaceFlow := replaceFlow
			replaceFlow = func (string, logger.LoggerInterface, string) error {
				oldreplaceFlowNumCalls++
				return nil
			}
			
			err := ValidateFlows(testLogger, basedir)
			assert.NoError(t, err)

			assert.Equal(t, 1, IsFlowValidCalls)
			assert.Equal(t, 0, oldreplaceFlowNumCalls)

			replaceFlow = oldreplaceFlow
		})

		t.Run("When IsFlowValid returns false", func (t *testing.T) {
			// setup the toolkit flow every time, since it gets deleted at the end of the function
			err = os.Mkdir(basedir + string(os.PathSeparator) + "temp", 0777)
			assert.NoError(t, err)

			IsFlowValidCalls := 0
			IsFlowValid = func (logger.LoggerInterface, string, []byte) (bool, error) {
				IsFlowValidCalls++;
				return false, nil
			}

			t.Run("When replaceFlow fails", func (t *testing.T) {
				IsFlowValidCalls = 0
				oldreplaceFlowNumCalls := 0
				oldreplaceFlow := replaceFlow
				replaceFlow = func (string, logger.LoggerInterface, string) error {
					oldreplaceFlowNumCalls++
					return errors.New("Test")
				}	

				err := ValidateFlows(testLogger, basedir)
				assert.Error(t, err)

				assert.Equal(t, 1, IsFlowValidCalls)
				assert.Equal(t, 1, oldreplaceFlowNumCalls)
				replaceFlow = oldreplaceFlow
			})

			t.Run("When replaceFlow does not fail", func (t *testing.T) {
				IsFlowValidCalls = 0
				oldreplaceFlowNumCalls := 0
				oldreplaceFlow := replaceFlow
				replaceFlow = func (string, logger.LoggerInterface, string) error {
					oldreplaceFlowNumCalls++
					return nil
				}

				err := ValidateFlows(testLogger, basedir)
				assert.NoError(t, err)

				assert.Equal(t, 1, IsFlowValidCalls)
				assert.Equal(t, 1, oldreplaceFlowNumCalls)
				replaceFlow = oldreplaceFlow
			})
		})

		IsFlowValid = oldIsFlowValid

		// cleanup folders	
		err = os.RemoveAll(basedir + string(os.PathSeparator) + "ace-server")
		assert.NoError(t, err)
	})
}

var flowDocumentYAML string = `$integration: 'http://ibm.com/appconnect/integration/v2/integrationFile'
integration:
  type: api
  trigger-interfaces:
    trigger-interface-1:
      triggers:
        retrieveTest:
          assembly:
            $ref: '#/integration/assemblies/assembly-1'
          input-context:
            data: test
          output-context:
            data: test
      options:
        resources:
          - business-object: test
            model:
              $ref: '#/models/test'
            triggers:
              retrieve: retrieveTest
      type: api-trigger
  action-interfaces:
    action-interface-1:
      type: api-action
      business-object: mail
      connector-type: gmail
      account-name: Account 1
      actions:
        CREATE: {}
  assemblies:
    assembly-1:
      assembly:
        execute:
          - create-action:
              name: Gmail Create email
              target:
                $ref: '#/integration/action-interfaces/action-interface-1'
              map:
                mappings:
                  - To:
                      template: test@gmail.com
                $map: 'http://ibm.com/appconnect/map/v1'
                input:
                  - variable: api
                    $ref: '#/trigger/api/parameters'
          - response:
              name: response-1
              reply-maps:
                - title: test successfully retrieved
                  status-code: '200'
                  map:
                    $map: 'http://ibm.com/appconnect/map/v1'
                    input:
                      - variable: api
                        $ref: '#/trigger/api/parameters'
                      - variable: GmailCreateemail
                        $ref: '#/node-output/Gmail Create email/response/payload'
                    mappings: []
                  input:
                    - variable: api
                      $ref: '#/trigger/api/parameters'
                    - variable: GmailCreateemail
                      $ref: '#/node-output/Gmail Create email/response/payload'
  name: Untitled API 1
models: {}`