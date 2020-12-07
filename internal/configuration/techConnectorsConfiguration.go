package configuration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/aymerick/raymond"
	"github.com/ghodss/yaml"
	"github.com/ot4i/ace-docker/internal/logger"
)

// Credentials structure for jdbc credentials object, can be extended for other tech connectors
type Credentials struct {
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

// AccountInfo structure for individual account
type AccountInfo struct {
	Name        string      `json:"name"`
	Credentials Credentials `json:"credentials"`
}

// Accounts connectos account object
type Accounts struct {
	Accounts map[string][]AccountInfo `json:"accounts"`
}

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

func setDSNForJDBCApplication(log *logger.Logger, basedir string, jsonContentsObjForCredParse *Accounts) error {
	log.Println("#setDSNForJDBCApplication: Execute mqsisetdbparms command	")
	if _, ok := jsonContentsObjForCredParse.Accounts["jdbc"]; ok {
		jdbcAccountsToGetActCred := jsonContentsObjForCredParse.Accounts["jdbc"]
		for _, accountContent := range jdbcAccountsToGetActCred {
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
	}
	return nil
}

// SetupTechConnectorsConfigurations entry point for all technology connector configurations
func SetupTechConnectorsConfigurations(log *logger.Logger, basedir string, contents []byte) error {
	log.Println("#SetupTechConnectorsConfigurations: extracting accounts info")
	jsonContents, err := yaml.YAMLToJSON(contents)
	if err != nil {
		log.Printf("#SetupTechConnectorsConfigurations YAMLToJSON: %v\n", err)
		return nil
	}
	var jsonContentsObjForCredParse Accounts
	err = json.Unmarshal(jsonContents, &jsonContentsObjForCredParse)
	if err != nil {
		log.Fatalf("#SetupTechConnectorsConfigurations Unmarshal: %v", err)
	}

	if jsonContentsObjForCredParse.Accounts != nil {
		log.Println("#SetupTechConnectorsConfigurations: checking each tech connector configuration")

		if _, ok := jsonContentsObjForCredParse.Accounts["jdbc"]; ok {
			err = setDSNForJDBCApplication(log, basedir, &jsonContentsObjForCredParse)

			if err != nil {
				log.Printf("#SetupTechConnectorsConfigurations: encountered an error in setDSNForJDBCApplication: %v\n", err)
				return err
			}
			err = buildJDBCPolicies(log, basedir, &jsonContentsObjForCredParse)
			if err != nil {
				log.Printf("#SetupTechConnectorsConfigurations: encountered an error in buildJDBCPolicies: %v\n", err)
				return err
			}
		}
	}
	return nil
}

func buildJDBCPolicies(log *logger.Logger, basedir string, jsonContentsObjForCredParse *Accounts) error {
	var supportedDBs = map[string]string{
		"IBM Db2 Linux, UNIX, or Windows (LUW) - client managed": "db2luw",
		"IBM Db2 Linux, UNIX, or Windows (LUW) - IBM Cloud":      "db2cloud",
		"IBM Db2 for i": "db2i",
		"Oracle":        "oracle",
		"PostgreSQL":    "postgresql",
	}

	jdbcAccounts := jsonContentsObjForCredParse.Accounts["jdbc"]

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

	if _, err := os.Stat(policyDirName); os.IsNotExist(err) {
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

		tpl, err := raymond.Parse(string(policyTemplate))
		if err != nil {
			log.Printf("#buildJDBCPolicies: failed to parse the template - %v", err)
			return err
		}

		result, err := tpl.Exec(context)
		if err != nil {
			log.Printf("#buildJDBCPolicies: rendering failed with an error - %v", err)
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

func getJDBCPolicyAttributes(log *logger.Logger, dbType, hostname, port, dbName, additonalParams string) (map[string]string, error) {

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
