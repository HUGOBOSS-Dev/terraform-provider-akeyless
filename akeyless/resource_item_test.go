package akeyless

import (
	"context"
	"fmt"
	"testing"

	"github.com/akeylesslabs/akeyless-go/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestDfcKeyRsaResource(t *testing.T) {
	t.Parallel()

	cert := generateCertForTest(t, 1024)

	name := "test_rsa_key"
	itemPath := testPath(name)

	config := fmt.Sprintf(`
		resource "akeyless_dfc_key" "%v" {
			name 								= "%v"
			alg 								= "RSA1024"
			description 						= "aaaa"
			split_level 						= 2
			generate_self_signed_certificate 	= true
			certificate_ttl 					= 60
			certificate_common_name 			= "cn1"
			certificate_organization 			= "org1"
			certificate_country 				= "cntry1"
			certificate_locality 				= "local1"
			certificate_province 				= "prov1"
			tags        						= ["t1","t2"]
			delete_protection 					= true
		}
	`, name, itemPath)

	configUpdate := fmt.Sprintf(`
		resource "akeyless_dfc_key" "%v" {
			name 								= "%v"
			alg 								= "RSA1024"
			description 						= "bbbb"
			split_level 						= 2
			generate_self_signed_certificate 	= true
			certificate_ttl 					= 60
			certificate_common_name 			= "cn1"
			certificate_organization 			= "org1"
			certificate_country 				= "cntry1"
			certificate_locality 				= "local1"
			certificate_province 				= "prov1"
			cert_data_base64 					= "%v"
			tags        						= ["t1","t3"]
		}
	`, name, itemPath, cert)

	tesItemResource(t, config, configUpdate, itemPath)
}

func TestDfcKeyAesResource(t *testing.T) {
	t.Parallel()

	name := "test_dfc_key"
	itemPath := testPath("path_dfc_key12")
	config := fmt.Sprintf(`
		resource "akeyless_dfc_key" "%v" {
			name = "%v"
			tags     = ["t1", "t2"]
			alg = "AES128SIV"
		}
	`, name, itemPath)

	configUpdate := fmt.Sprintf(`
		resource "akeyless_dfc_key" "%v" {
			name = "%v"	
			tags     = ["t1", "t3"]
			alg = "AES128SIV"
		}
	`, name, itemPath)

	tesItemResource(t, config, configUpdate, itemPath)
}

func TestRsaPublicResource(t *testing.T) {
	t.Parallel()

	name := "test_rsa_pub_key"
	itemPath := testPath(name)

	config := fmt.Sprintf(`
		resource "akeyless_dfc_key" "%v" {
			name = "%v"
			alg = "RSA2048"
		}
		data "akeyless_rsa_pub" "%v_2" {
			name = akeyless_dfc_key.%v.name
		}
	`, name, itemPath, name, name)

	configUpdate := config

	tesItemResource(t, config, configUpdate, itemPath)
}

func TestClassicKey(t *testing.T) {

	t.Skip("not authorized to create producer on public gateway")
	t.Parallel()

	name := "test_classic_key"
	itemPath := testPath(name)

	config := fmt.Sprintf(`
		resource "akeyless_classic_key" "%v" {
			name 		= "%v"
			alg 		= "RSA2048"
			tags 		= ["aaaa", "bbbb"]
		}
	`, name, itemPath)

	configUpdate := fmt.Sprintf(`
		resource "akeyless_classic_key" "%v" {
			name 		= "%v"	
			alg 		= "RSA2048"
			tags 		= ["cccc", "dddd"]
			metadata 	= "abcd"
		}
	`, name, itemPath)

	tesItemResource(t, config, configUpdate, itemPath)
}

func TestPkiResource(t *testing.T) {
	t.Parallel()

	keyPath := testPath("test-dfc-for-pki")
	createDfcKey(t, keyPath)
	defer deleteItem(t, keyPath)

	name := "test-pki-resource"
	itemPath := testPath(name)

	config := fmt.Sprintf(`
		resource "akeyless_pki_cert_issuer" "%v" {
			name 					= "%v"
			signer_key_name 		= "/%v"
			ttl                   	= 60
			gw_cluster_url        	= "http://localhost:8000"
			destination_path      	= "/terraform-tests"
			allowed_domains       	= "domains"
			allowed_uri_sans      	= "uri_sans"
			allow_subdomains      	= true
			not_enforce_hostnames 	= false
			allow_any_name        	= true
			not_require_cn        	= true
			server_flag           	= true
			client_flag           	= true
			code_signing_flag     	= true
			key_usage             	= "KeyAgreement,KeyEncipherment"
			organizational_units  	= "org1"
			country               	= "coun1"
			locality              	= "loca1"
			province              	= "prov1"
			street_address        	= "stre1"
			postal_code           	= "post1"
			protect_certificates  	= true
			expiration_event_in   	= ["1"]
			description           	= "desc1"
			tags     			  	= ["t1", "t2"]
			delete_protection     	= "true"
		}
	`, name, itemPath, keyPath)

	configUpdate := fmt.Sprintf(`
		resource "akeyless_pki_cert_issuer" "%v" {
			name 					= "%v"
			signer_key_name 		= "/%v"
			ttl                   	= 90
			gw_cluster_url        	= "http://localhost:8000"
			destination_path      	= "/terraform-tests"
			allowed_domains       	= "domain1,domain2"
			allowed_uri_sans      	= "uri_san1,uri_san2"
			allow_subdomains      	= false
			not_enforce_hostnames 	= true
			allow_any_name        	= false
			not_require_cn        	= true
			server_flag           	= false
			client_flag           	= false
			code_signing_flag     	= false
			key_usage             	= "DigitalSignature"
			organizational_units  	= "org1,org2"
			country               	= "coun2"
			locality              	= "loca2"
			province              	= "prov2"
			street_address        	= "stre2"
			postal_code           	= "post2"
			protect_certificates  	= false
			expiration_event_in   	= []
			description           	= "desc2"
			tags     			  	= ["t1", "t3"]
		}
	`, name, itemPath, keyPath)

	tesItemResource(t, config, configUpdate, itemPath)
}

func TestPkiDataSource(t *testing.T) {
	t.Parallel()

	privateKey, csr := generateKeyAndCsrForTest(1024)

	// create key
	keyName := "test-dfc-for-pki-test"
	keyPath := testPath(keyName)
	createDfcKey(t, keyPath)
	defer deleteItem(t, keyPath)

	// create pki-cert-issuer
	name := "test-pki-data"
	itemPath := testPath(name)
	destPath := "terraform-tests"
	cn := "cn1"
	uriSan := "uri1"
	createPkiCertIssuer(t, keyPath, itemPath, destPath, cn, uriSan)
	defer deleteItem(t, itemPath)

	// pki certificates must be deleted before deleting the pki issuer on cleanup
	certsPath := fmt.Sprintf("/%s/%s", destPath, cn)
	defer deleteItems(t, certsPath)

	// with key
	config1 := fmt.Sprintf(`
		data "akeyless_pki_certificate" "pki_cert" {
			cert_issuer_name  	= "%v"
			key_data_base64   	= "%v"
			common_name         = "%v"
			alt_names           = "%v"
			uri_sans            = "%v"
			ttl                 = 120
			extended_key_usage  = "clientauth"
		}
		output "pki" {
			value     = data.akeyless_pki_certificate.pki_cert
			sensitive = true
		}
	`, itemPath, privateKey, cn, cn, uriSan)

	tesItemDataSource(t, config1, "pki", []string{"data", "parent_cert"})

	// with csr
	config2 := fmt.Sprintf(`
		data "akeyless_pki_certificate" "pki_cert" {
			cert_issuer_name  	= "%v"
			csr_data_base64     = "%v"
			common_name         = "%v"
			ttl                 = 120
			extended_key_usage  = "clientauth"
		}
		output "pki" {
			value     = data.akeyless_pki_certificate.pki_cert
			sensitive = true
		}
	`, itemPath, csr, cn)

	tesItemDataSource(t, config2, "pki", []string{"data", "parent_cert"})
}

func TestSshCertResource(t *testing.T) {
	t.Parallel()

	name := "test_ssh"
	itemPath := testPath(name)

	config := fmt.Sprintf(`
		resource "akeyless_dfc_key" "key_ssh" {
			name = "terraform-tests/test_ssh_key"
			alg = "RSA1024"
		}
		resource "akeyless_ssh_cert_issuer" "%v" {
			name 							= "%v"
			ttl 							= "500"
			signer_key_name 				= "/terraform-tests/test_ssh_key"
			tags     						= ["t1", "t2"]
			allowed_users 					= "aaaa"
			secure_access_enable 			= "true"
			secure_access_host 				= ["1.1.1.1", "2.2.2.2"]
			secure_access_bastion_api 		= "https://my.bastion:9900"
			secure_access_bastion_ssh 		= "my.bastion:22"
			secure_access_ssh_creds_user 	= "aaaa"
			delete_protection 				= true

			depends_on = [
    			akeyless_dfc_key.key_ssh,
  			]
		}
	`, name, itemPath)

	configUpdate := fmt.Sprintf(`
		resource "akeyless_dfc_key" "key_ssh" {
			name = "terraform-tests/test_ssh_key"
			alg = "RSA1024"
			tags     = ["t1", "t2"]
		}

		resource "akeyless_ssh_cert_issuer" "%v" {
			name = "%v"
			ttl = "290"
			signer_key_name = "/terraform-tests/test_ssh_key"
			tags     = ["t1", "t3"]
			allowed_users = "aaaa2,fffff"
			secure_access_enable = "true"
			secure_access_host = ["1.1.1.1", "2.2.2.2"]
			secure_access_bastion_api = "https://my.bastion:9901"
			secure_access_bastion_ssh = "my.bastion1:22"
			secure_access_ssh_creds_user = "aaaa2"

			depends_on = [
    			akeyless_dfc_key.key_ssh,
  			]
		}
	`, name, itemPath)

	tesItemResource(t, config, configUpdate, itemPath)
}

func TestSshDataSource(t *testing.T) {
	t.Parallel()

	// create key
	keyName := "test-dfc-for-ssh-test"
	keyPath := testPath(keyName)
	createDfcKey(t, keyPath)
	defer deleteItem(t, keyPath)

	rsaPublicKey := getRsaPublicKey(t, keyPath)
	sshPublicKey := *rsaPublicKey.Ssh

	// create ssh-cert-issuer
	name := "test-ssh-data"
	itemPath := testPath(name)
	allowedUser := "tf_user"
	createSshCertIssuer(t, keyPath, itemPath, allowedUser)
	defer deleteItem(t, itemPath)

	config1 := fmt.Sprintf(`
		data "akeyless_ssh_certificate" "ssh_cert" {
			cert_issuer_name  		= "%v"
			cert_username     		= "%v"
			public_key_data   		= "%v"
			ttl 					= 120
		}
		output "ssh" {
			value     = data.akeyless_ssh_certificate.ssh_cert
			sensitive = true
		}
	`, itemPath, allowedUser, sshPublicKey)

	tesItemDataSource(t, config1, "ssh", []string{"data"})

	config2 := fmt.Sprintf(`
		data "akeyless_ssh_certificate" "ssh_cert" {
			cert_issuer_name  		= "%v"
			cert_username     		= "%v"
			public_key_data   		= "%v"
			ttl 					= 180
			legacy_signing_alg_name = true
		}
		output "ssh" {
			value     = data.akeyless_ssh_certificate.ssh_cert
			sensitive = true
		}
	`, itemPath, allowedUser, sshPublicKey)

	tesItemDataSource(t, config2, "ssh", []string{"data"})
}

func tesItemResource(t *testing.T, config, configUpdate, itemPath string) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				//PreConfig: deleteFunc,
				Check: resource.ComposeTestCheckFunc(
					checkItemExistsRemotely(itemPath),
				),
			},
			{
				Config: configUpdate,
				Check: resource.ComposeTestCheckFunc(
					checkItemExistsRemotely(itemPath),
				),
			},
		},
	})
}

func checkItemExistsRemotely(path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := *testAccProvider.Meta().(providerMeta).client
		token := *testAccProvider.Meta().(providerMeta).token

		gsvBody := akeyless.DescribeItem{
			Name:         path,
			ShowVersions: akeyless.PtrBool(false),
			Token:        &token,
		}

		_, _, err := client.DescribeItem(context.Background()).Body(gsvBody).Execute()
		if err != nil {
			return err
		}
		return nil
	}
}

func tesItemDataSource(t *testing.T, config, outputName string, params []string) {

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkOutputNotEmpty(t, outputName, params),
			},
		},
	})
}

func checkOutputNotEmpty(t *testing.T, name string, params []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		outputs, ok := ms.Outputs[name]
		if !ok || outputs == nil {
			return nil
		}
		values := outputs.Value.(map[string]interface{})

		for _, param := range params {
			rs, ok := values[param]
			if !ok {
				return fmt.Errorf("output '%s' not found", param)
			}
			output, ok := rs.(string)
			if !ok || output == "" {
				return fmt.Errorf("output '%s' not found", param)
			}
		}
		return nil
	}
}
