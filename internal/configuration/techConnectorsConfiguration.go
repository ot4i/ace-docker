package configuration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/ghodss/yaml"
	"github.com/ot4i/ace-docker/common/logger"
)

// JdbcCredentials Credentials structure for jdbc credentials object, can be extended for other tech connectors
type JdbcCredentials struct {
	AuthType        interface{} `json:"authType"`
	DbType          interface{} `json:"dbType"`
	Hostname        interface{} `json:"hostname"`
	Port            interface{} `json:"port"`
	DbName          interface{} `json:"dbName"`
	Username        interface{} `json:"username"`
	Password        interface{} `json:"password"`
	MaxPoolSize     interface{} `json:"maxPoolSize"`
	AdditonalParams interface{} `json:"additonalParams"`
}

// MQCredentials Credentials structure for mq account credentials object
type MQCredentials struct {
	AuthType     string      `json:"authType"`
	QueueManager string      `json:"queueManager"`
	Hostname     string      `json:"hostname"`
	Port         interface{} `json:"port"`
	Username     string      `json:"username"`
	Password     string      `json:"password"`
	ChannelName  string      `json:"channelName"`
}

// JdbcAccountInfo structure for jdbc connector accounts
type JdbcAccountInfo struct {
	Name        string
	Credentials JdbcCredentials
}

// MQAccountInfo Account info structure mq connector accounts
type MQAccountInfo struct {
	Name        string
	Credentials MQCredentials
}

// AccountInfo structure for individual account
type AccountInfo struct {
	Name        string          `json:"name"`
	Credentials json.RawMessage `json:"credentials"`
}

// Accounts connectos account object
type Accounts struct {
	Accounts map[string][]AccountInfo `json:"accounts"`
}

var processMqConnectorAccounts = processMQConnectorAccountsImpl
var processJdbcConnectorAccounts = processJdbcConnectorAccountsImpl
var raymondParse = raymond.Parse

func convertToString(unknown interface{}) string {
	switch unknown.(type) {
	case float64:
		return fmt.Sprintf("%v", unknown)
	case string:
		return unknown.(string)
	default:
		return ""
	}
}

func convertToNumber(unknown interface{}) float64 {
	switch i := unknown.(type) {
	case float64:
		return i
	case string:
		f, _ := strconv.ParseFloat(i, 64)
		return f
	default:
		return 0
	}
}

// SetupTechConnectorsConfigurations entry point for all technology connector configurations
func SetupTechConnectorsConfigurations(log logger.LoggerInterface, basedir string, contents []byte) error {

	techConnectors := map[string]func(log logger.LoggerInterface, basedir string, accounts []AccountInfo) error{
		"jdbc": processJdbcConnectorAccounts,
		"mq":   processMqConnectorAccounts}

	log.Println("#SetupTechConnectorsConfigurations: extracting accounts info")
	jsonContents, err := yaml.YAMLToJSON(contents)
	if err != nil {
		log.Printf("#SetupTechConnectorsConfigurations YAMLToJSON: %v\n", err)
		return err
	}

	var jsonContentsObjForCredParse Accounts
	err = json.Unmarshal(jsonContents, &jsonContentsObjForCredParse)
	if err != nil {
		log.Fatalf("#SetupTechConnectorsConfigurations Unmarshal: %v", err)
		return nil
	}

	for connector, connectorFunc := range techConnectors {
		connectorAccounts := jsonContentsObjForCredParse.Accounts[connector]

		if len(connectorAccounts) > 0 {
			log.Printf("Processing connector %s accounts \n", connector)
			err := connectorFunc(log, basedir, connectorAccounts)

			if err != nil {
				log.Printf("An error occured while proccessing connector accounts %s %v\n", connector, err)
				return err
			} else {
				log.Printf("Connector %s accounts processed %v", connector, len(connectorAccounts))
			}
		} else {
			log.Printf("No accounts found for connector %s\n", connector)
		}
	}

	return nil
}

func processJdbcConnectorAccountsImpl(log logger.LoggerInterface, basedir string, accounts []AccountInfo) error {

	designerAuthMode, ok := os.LookupEnv("DEVELOPMENT_MODE")

	if ok && designerAuthMode == "true" {
		log.Println("Ignore jdbc accounts in designer authoring integration server")
		return nil
	}

	jdbcAccounts := unmarshalJdbcAccounts(accounts)
	err := setDSNForJDBCApplication(log, basedir, jdbcAccounts)

	if err != nil {
		log.Printf("#SetupTechConnectorsConfigurations: encountered an error in setDSNForJDBCApplication: %v\n", err)
		return err
	}
	err = buildJDBCPolicies(log, basedir, jdbcAccounts)
	if err != nil {
		log.Printf("#SetupTechConnectorsConfigurations: encountered an error in buildJDBCPolicies: %v\n", err)
		return err
	}

	return nil
}

var setDSNForJDBCApplication = func(log logger.LoggerInterface, basedir string, jdbcAccounts []JdbcAccountInfo) error {
	log.Println("#setDSNForJDBCApplication: Execute mqsisetdbparms command	")

	for _, accountContent := range jdbcAccounts {
		log.Printf("#setDSNForJDBCApplication: setting up config for account - %v\n", accountContent.Name)
		jdbcCurrAccountCredInfo := accountContent.Credentials

		hostName := convertToString(jdbcCurrAccountCredInfo.Hostname)
		dbPort := convertToString(jdbcCurrAccountCredInfo.Port)
		dbName := convertToString(jdbcCurrAccountCredInfo.DbName)
		userName := convertToString(jdbcCurrAccountCredInfo.Username)
		password := convertToString(jdbcCurrAccountCredInfo.Password)

		if len(hostName) == 0 || len(dbPort) == 0 || len(dbName) == 0 || len(userName) == 0 || len(password) == 0 {
			log.Printf("#setDSNForJDBCApplication: skipping executing mqsisetdbparms for account - %v as one of the required fields found empty\n", accountContent.Name)
			continue
		}
		shaInputRawText := hostName + ":" + dbPort + ":" + dbName + ":" + userName
		hash := sha256.New()
		hash.Write([]byte(shaInputRawText))
		shaHashEncodedText := hex.EncodeToString(hash.Sum(nil))
		args := []string{"'-n'", "jdbc::" + shaHashEncodedText, "'-u'", userName, "'-p'", password, "'-w'", "'" + basedir + string(os.PathSeparator) + workdirName + "'"}
		err := internalRunSetdbparmsCommand(log, "mqsisetdbparms", args)
		if err != nil {
			return err
		}
	}
	return nil
}

func unmarshalJdbcAccounts(accounts []AccountInfo) []JdbcAccountInfo {

	jdbcAccountsInfo := make([]JdbcAccountInfo, len(accounts))

	for i, accountInfo := range accounts {
		jdbcAccountsInfo[i].Name = accountInfo.Name
		json.Unmarshal(accountInfo.Credentials, &jdbcAccountsInfo[i].Credentials)

	}

	return jdbcAccountsInfo
}

var buildJDBCPolicies = func(log logger.LoggerInterface, basedir string, jdbcAccounts []JdbcAccountInfo) error {
	var supportedDBs = map[string]string{
		"IBM Db2 Linux, UNIX, or Windows (LUW) - client managed": "db2luw",
		"IBM Db2 Linux, UNIX, or Windows (LUW) - IBM Cloud":      "db2cloud",
		"IBM Db2 for i": "db2i",
		"Oracle":        "oracle",
		"PostgreSQL":    "postgresql",
	}

	policyDirName := basedir + string(os.PathSeparator) + workdirName + string(os.PathSeparator) + "overrides" + string(os.PathSeparator) + "gen.jdbcConnectorPolicies"
	log.Printf("#buildJDBCPolicies: jdbc policy directory  %v\n", policyDirName)

	policyNameSuffix := ".policyxml"

	policyxmlDescriptor := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
   <ns2:policyProjectDescriptor xmlns="http://com.ibm.etools.mft.descriptor.base" xmlns:ns2="http://com.ibm.etools.mft.descriptor.policyProject">
	  <references/>
	</ns2:policyProjectDescriptor>`

	policyTemplate := `<?xml version="1.0" encoding="UTF-8"?>
 <policies>
  <policy policyType="JDBCProviders" policyName="{{{policyName}}}" policyTemplate="DB2_91_Linux">
   <databaseName>{{{dbName}}}</databaseName>
   <databaseType>{{{dbType}}}</databaseType>
   <databaseVersion></databaseVersion>
   <type4DriverClassName>{{{jdbcClassName}}}</type4DriverClassName>
   <type4DatasourceClassName>{{{jdbcType4DataSourceName}}}</type4DatasourceClassName>
   <connectionUrlFormat>{{{jdbcURL}}}</connectionUrlFormat>
   <connectionUrlFormatAttr1></connectionUrlFormatAttr1>
   <connectionUrlFormatAttr2></connectionUrlFormatAttr2>
   <connectionUrlFormatAttr3></connectionUrlFormatAttr3>
   <connectionUrlFormatAttr4></connectionUrlFormatAttr4>
   <connectionUrlFormatAttr5></connectionUrlFormatAttr5>
   <serverName>{{{hostname}}}</serverName>
   <portNumber>{{{port}}}</portNumber>
   <jarsURL></jarsURL>
   <databaseSchemaNames>useProvidedSchemaNames</databaseSchemaNames>
   <description></description>
   <maxConnectionPoolSize>{{{maxPoolSize}}}</maxConnectionPoolSize>
   <securityIdentity>{{{securityIdentity}}}</securityIdentity>
   <environmentParms></environmentParms>
   <jdbcProviderXASupport>false</jdbcProviderXASupport>
   <useDeployedJars>true</useDeployedJars>
 </policy>
</policies>
`

	if _, err := osStat(policyDirName); osIsNotExist(err) {
		log.Printf("#buildJDBCPolicies: %v does not exist, creating afresh..", policyDirName)
		err := osMkdirAll(policyDirName, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		log.Printf("#buildJDBCPolicies: %v already exists", policyDirName)
	}

	for _, accountContent := range jdbcAccounts {
		log.Printf("#buildJDBCPolicies: building policy for account : \"%v\"", accountContent.Name)
		jdbcAccountInfo := accountContent.Credentials

		dbType := convertToString(jdbcAccountInfo.DbType)
		hostName := convertToString(jdbcAccountInfo.Hostname)
		dbPort := convertToString(jdbcAccountInfo.Port)
		dbName := convertToString(jdbcAccountInfo.DbName)
		userName := convertToString(jdbcAccountInfo.Username)
		password := convertToString(jdbcAccountInfo.Password)
		additionalParams := convertToString(jdbcAccountInfo.AdditonalParams)

		if len(hostName) == 0 || len(dbPort) == 0 || len(dbName) == 0 || len(userName) == 0 || len(password) == 0 {
			log.Printf("#buildJDBCPolicies: skipping building policy for account - %v as one of the required fields found empty\n", accountContent.Name)
			continue
		}

		rawText := hostName + ":" + dbPort + ":" + dbName + ":" + userName
		hash := sha256.New()
		hash.Write([]byte(rawText))
		uuid := hex.EncodeToString(hash.Sum(nil))

		databaseType := supportedDBs[dbType]

		policyAttributes, err := getJDBCPolicyAttributes(log, databaseType, hostName, dbPort, dbName, additionalParams)
		if err != nil {
			log.Printf("#buildJDBCPolicies: getJDBCPolicyAttributes returned an error - %v", err)
			return err
		}

		policyName := databaseType + "-" + uuid
		context := map[string]interface{}{
			"policyName":              policyName,
			"dbName":                  dbName,
			"dbType":                  databaseType,
			"jdbcClassName":           policyAttributes["jdbcClassName"],
			"jdbcType4DataSourceName": policyAttributes["jdbcType4DataSourceName"],
			"jdbcURL":                 policyAttributes["jdbcURL"],
			"hostname":                hostName,
			"port":                    convertToNumber(jdbcAccountInfo.Port),
			"maxPoolSize":             convertToNumber(jdbcAccountInfo.MaxPoolSize),
			"securityIdentity":        uuid,
		}

		result, err := transformXMLTemplate(string(policyTemplate), context)
		if err != nil {
			log.Printf("#buildJDBCPolicies: failed to transform policy xml - %v", err)
			return err
		}

		policyFileName := policyName + policyNameSuffix

		err = ioutilWriteFile(policyDirName+string(os.PathSeparator)+policyFileName, []byte(result), os.ModePerm)
		if err != nil {
			log.Printf("#buildJDBCPolicies: failed to write to the policy file %v - %v", policyFileName, err)
			return err
		}
	}

	err := ioutilWriteFile(policyDirName+string(os.PathSeparator)+"policy.descriptor", []byte(policyxmlDescriptor), os.ModePerm)
	if err != nil {
		log.Printf("#buildJDBCPolicies: failed to write to the policy.descriptor - %v", err)
		return err
	}
	return nil
}

func getJDBCPolicyAttributes(log logger.LoggerInterface, dbType, hostname, port, dbName, additonalParams string) (map[string]string, error) {

	var policyAttributes = make(map[string]string)
	var classNames = map[string]string{
		"DB2NativeDriverClassName":     "com.ibm.db2.jcc.DB2Driver",
		"DB2NativeDataSourceClassName": "com.ibm.db2.jcc.DB2XADataSource",
		"DB2DriverClassName":           "com.ibm.appconnect.jdbc.db2.DB2Driver",
		"DB2DataSourceClassName":       "com.ibm.appconnect.jdbcx.db2.DB2DataSource",
		"OracleDriverClassName":        "com.ibm.appconnect.jdbc.oracle.OracleDriver",
		"OracleDataSourceClassName":    "com.ibm.appconnect.jdbcx.oracle.OracleDataSource",
		"MySQLDriverClassName":         "com.ibm.appconnect.jdbc.mysql.MySQLDriver",
		"MySQLDataSourceClassName":     "com.ibm.appconnect.jdbcx.mysql.MySQLDataSource",
		"SqlServerDriverClassName":     "com.ibm.appconnect.jdbc.sqlserver.SQLServerDriver",
		"SqlServerDataSourceClassName": "com.ibm.appconnect.jdbcx.sqlserver.SQLServerDataSource",
		"PostgresDriveClassName":       "com.ibm.appconnect.jdbc.postgresql.PostgreSQLDriver",
		"PostgresDataSourceClassName":  "com.ibm.appconnect.jdbcx.postgresql.PostgreSQLDataSource",
		"HiveDriverClassName":          "com.ibm.appconnect.jdbc.hive.HiveDriver",
		"HiveDataSourceClassName":      "com.ibm.appconnect.jdbcx.hive.HiveDataSource",
	}

	var jdbcURL, jdbcClassName, jdbcType4DataSourceName string
	var endDemiliter = ""

	var err error
	switch dbType {
	case "db2luw", "db2cloud":
		jdbcURL = "jdbc:db2://" + hostname + ":" + port + "/" + dbName + ":user=[user];password=[password];loginTimeout=40"
		jdbcClassName = classNames["DB2NativeDriverClassName"]
		jdbcType4DataSourceName = classNames["DB2NativeDataSourceClassName"]
		endDemiliter = ";"
	case "db2i":
		jdbcURL = "jdbc:ibmappconnect:db2://" + hostname + ":" + port + ";DatabaseName=" + dbName + ";user=[user];password=[password];loginTimeout=40"
		jdbcClassName = classNames["DB2DriverClassName"]
		jdbcType4DataSourceName = classNames["DB2DataSourceClassName"]
	case "oracle":
		jdbcURL = "jdbc:ibmappconnect:oracle://" + hostname + ":" + port + ";DatabaseName=" + dbName + ";user=[user];password=[password];loginTimeout=40;FetchDateAsTimestamp=false"
		jdbcClassName = classNames["OracleDriverClassName"]
		jdbcType4DataSourceName = classNames["OracleDataSourceClassName"]
	case "postgresql":
		jdbcURL = "jdbc:ibmappconnect:postgresql://" + hostname + ":" + port + ";DatabaseName=" + dbName + ";user=[user];password=[password];loginTimeout=40"
		jdbcClassName = classNames["PostgresDriveClassName"]
		jdbcType4DataSourceName = classNames["PostgresDataSourceClassName"]
	default:
		err = errors.New("Unsupported database type: " + dbType)
		return nil, err
	}

	if additonalParams != "" {
		jdbcURL = jdbcURL + ";" + additonalParams
	}

	if endDemiliter != "" {
		jdbcURL += endDemiliter
	}

	policyAttributes["jdbcURL"] = jdbcURL
	policyAttributes["jdbcClassName"] = jdbcClassName
	policyAttributes["jdbcType4DataSourceName"] = jdbcType4DataSourceName
	return policyAttributes, err
}

func haveMultipleBarFiles() bool {
	urls := os.Getenv("ACE_CONTENT_SERVER_URL")
	urlArray := strings.Split(urls, ",")

	return len(urlArray) > 1
}

var processMQConnectorAccountsImpl = func(log logger.LoggerInterface, basedir string, accounts []AccountInfo) error {

	mqAccounts := unmarshalMQAccounts(accounts)

	designerAuthMode, ok := os.LookupEnv("DEVELOPMENT_MODE")

	isDesignerAuthoringMode := false
	if ok && designerAuthMode == "true" {
		isDesignerAuthoringMode = true
	}

	for _, mqAccount := range mqAccounts {
		log.Printf("MQ account %v Q Manager %v", mqAccount.Name, mqAccount.Credentials.QueueManager)
		err := processMqAccount(log, basedir, mqAccount, isDesignerAuthoringMode)
		if err != nil {
			log.Printf("#SetupTechConnectorsConfigurations encountered an error while processing mq account %v\n", err)
			return err
		}
	}

	return nil
}

var unmarshalMQAccounts = func(accounts []AccountInfo) []MQAccountInfo {

	mqAccountsInfo := make([]MQAccountInfo, len(accounts))

	for i, accountInfo := range accounts {
		mqAccountsInfo[i].Name = accountInfo.Name
		json.Unmarshal(accountInfo.Credentials, &mqAccountsInfo[i].Credentials)

	}

	return mqAccountsInfo
}

var processMqAccount = func(log logger.LoggerInterface, baseDir string, mqAccount MQAccountInfo, isDesignerAuthoringMode bool) error {
	err := createMqAccountDbParams(log, baseDir, mqAccount)

	if err != nil {
		log.Println("#processMQAccounts create db params failed")
		return err
	}

	if isDesignerAuthoringMode {
		return nil
	}

	if haveMultipleBarFiles() {
		log.Println("#processMQAccounts IBM MQ Connector not supported for muliple bar files")
		return errors.New("IBM MQ Connector not supported for muliple bar files")
	}

	err = createMQPolicy(log, baseDir, mqAccount)
	if err != nil {
		log.Println("#processMQAccounts build mq policies failed")
		return err
	}

	err = createMQFlowBarOverridesProperties(log, baseDir, mqAccount)

	if err != nil {
		log.Println("#processMQAccounts create mq flow bar overrides failed")
		return err
	}

	return nil
}

func getMQAccountSHA(mqAccountInfo *MQAccountInfo) string {
	mqCredentials := mqAccountInfo.Credentials
	shaInputRawText := mqCredentials.Hostname + ":" + convertToString(mqCredentials.Port) + ":" + mqCredentials.QueueManager + ":" + mqCredentials.Username + ":" + mqCredentials.ChannelName
	hash := sha256.New()
	hash.Write([]byte(shaInputRawText))
	uuid := hex.EncodeToString(hash.Sum(nil))
	return uuid
}

var createMQPolicy = func(log logger.LoggerInterface, basedir string, mqAccount MQAccountInfo) error {

	policyDirName := basedir + string(os.PathSeparator) + workdirName + string(os.PathSeparator) + "overrides" + string(os.PathSeparator) + "gen.MQPolicies"
	log.Printf("#buildMQPolicyies: mq policy directory  %v\n", policyDirName)

	policyNameSuffix := ".policyxml"

	policyxmlDescriptor := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
   <ns2:policyProjectDescriptor xmlns="http://com.ibm.etools.mft.descriptor.base" xmlns:ns2="http://com.ibm.etools.mft.descriptor.policyProject">
	  <references/>
	</ns2:policyProjectDescriptor>`

	policyxmlTemplate := `<?xml version="1.0" encoding="UTF-8"?>
	<policies>
	  <policy policyType="MQEndpoint" policyName="{{{policyName}}}" policyTemplate="MQEndpoint">
		<connection>CLIENT</connection>
		<destinationQueueManagerName>{{{queueManager}}}</destinationQueueManagerName>
		<queueManagerHostname>{{{hostName}}}</queueManagerHostname>
		<listenerPortNumber>{{{port}}}</listenerPortNumber>
		<channelName>{{{channelName}}}</channelName>
		<securityIdentity>{{{securityIdentity}}}</securityIdentity>
		<useSSL>false</useSSL>
		<SSLPeerName></SSLPeerName>
		<SSLCipherSpec></SSLCipherSpec>
	  </policy>
	</policies>
`

	if _, err := osStat(policyDirName); osIsNotExist(err) {
		log.Printf("#createMQPolicy: %v does not exist, creating afresh..", policyDirName)
		err := osMkdirAll(policyDirName, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		log.Printf("#createMQPolicy: %v already exists", policyDirName)
	}

	log.Printf("#createMQPolicy: building policy for account : \"%v\"", mqAccount.Name)

	specialCharsRegEx, err := regexp.Compile("[^a-zA-Z0-9]")
	mqAccountName := convertToString(mqAccount.Name)

	if err != nil {
		log.Printf("#createMQPolicy:  Failed to compile regex")
		return err
	}

	policyName := specialCharsRegEx.ReplaceAllString(mqAccountName, "_")

	securityIdentity := ""

	if mqAccount.Credentials.Username == "" && mqAccount.Credentials.Password == "" {
		log.Println("#createMQPolicy - setting security identity empty")
	} else {
		securityIdentity = "gen_" + getMQAccountSHA(&mqAccount)
	}

	context := map[string]interface{}{
		"policyName":       policyName,
		"queueManager":     mqAccount.Credentials.QueueManager,
		"hostName":         mqAccount.Credentials.Hostname,
		"port":             mqAccount.Credentials.Port,
		"channelName":      mqAccount.Credentials.ChannelName,
		"securityIdentity": securityIdentity,
	}

	result, err := transformXMLTemplate(string(policyxmlTemplate), context)
	if err != nil {
		log.Printf("#createMQPolicy: transformXmlTemplate failed with an error - %v", err)
		return err
	}

	policyFileName := policyName + policyNameSuffix

	err = ioutilWriteFile(policyDirName+string(os.PathSeparator)+policyFileName, []byte(result), os.ModePerm)
	if err != nil {
		log.Printf("#createMQPolicy: failed to write to the policy file %v - %v", policyFileName, err)
		return err
	}

	err = ioutilWriteFile(policyDirName+string(os.PathSeparator)+"policy.descriptor", []byte(policyxmlDescriptor), os.ModePerm)
	if err != nil {
		log.Printf("#createMQPolicy: failed to write to the policy.descriptor - %v", err)
		return err
	}
	return nil
}

var createMqAccountDbParams = func(log logger.LoggerInterface, basedir string, mqAccount MQAccountInfo) error {

	if mqAccount.Credentials.Username == "" && mqAccount.Credentials.Password == "" {
		log.Println("#createMqAccountDbParams - skipping setdbparams empty credentials")
		return nil
	}

	log.Printf("#createMqAccountDbParams: setdbparams for account - %v\n", mqAccount.Name)

	securityIdentityName := "gen_" + getMQAccountSHA(&mqAccount)
	args := []string{"'-n'", "mq::" + securityIdentityName, "'-u'", "'" + mqAccount.Credentials.Username + "'", "'-p'", "'" + mqAccount.Credentials.Password + "'", "'-w'", "'" + basedir + string(os.PathSeparator) + workdirName + "'"}
	err := internalRunSetdbparmsCommand(log, "mqsisetdbparms", args)
	if err != nil {
		return err
	}

	return nil
}

var createMQFlowBarOverridesProperties = func(log logger.LoggerInterface, basedir string, mqAccount MQAccountInfo) error {
	log.Println("#createMQFlowBarOverridesProperties: Execute mqapplybaroverride command")

	barOverridesConfigDir := "/home/aceuser/initial-config/bar_overrides"

	if _, err := osStat(barOverridesConfigDir); osIsNotExist(err) {
		err = osMkdirAll(barOverridesConfigDir, os.ModePerm)
		if err != nil {
			log.Printf("#createMQFlowBarOverridesProperties Failed to create bar_overrides folder %v", err)
			return err
		}
	}

	specialCharsRegEx, err := regexp.Compile("[^a-zA-Z0-9]")

	if err != nil {
		log.Printf("#createMQFlowBarOverridesProperties failed to compile regex")
		return err
	}

	barPropertiesFileContent := ""

	log.Printf("#createMQFlowBarOverridesProperties: setting up config for account - %v\n", mqAccount.Name)

	mqAccountName := convertToString(mqAccount.Name)
	plainAccountName := specialCharsRegEx.ReplaceAllString(mqAccountName, "_")
	accountFlowName := "gen.mq_" + plainAccountName
	accountFlowEndPoint := "/_lcp-mq-connect_" + getMQAccountSHA(&mqAccount)
	httpURLProperty := accountFlowName + "#HTTPInput.URLSpecifier=" + accountFlowEndPoint

	barPropertiesFileContent += httpURLProperty + "\n"

	barPropertiesFilePath := "/home/aceuser/initial-config/bar_overrides/barfile.properties"
	err = internalAppendFile(barPropertiesFilePath, []byte(barPropertiesFileContent), 0644)

	if err != nil {
		log.Println("#createMQFlowBarOverridesProperties failed to append to barfile.properties")
		return err
	}

	return nil
}

var transformXMLTemplate = func(xmlTemplate string, context interface{}) (string, error) {
	tpl, err := raymondParse(string(xmlTemplate))
	if err != nil {
		log.Printf("#createMQPolicy: failed to parse the template - %v", err)
		return "", err
	}

	result, err := tpl.Exec(context)
	if err != nil {
		log.Printf("#createMQPolicy: rendering failed with an error - %v", err)
		return "", err
	}

	return result, nil
}
