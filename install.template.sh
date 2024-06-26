# DataKit install script for UNIX-like OS
# Wed Aug 11 11:35:28 CST 2021
# Author: tanb@jiagouyun.com

# https://stackoverflow.com/questions/19339248/append-line-to-etc-hosts-file-with-shell-script/37824076
# usage: updateHosts ip domain1 domain2 domain3 ...
updateHosts() {
	for n in "$@"
	do
		if [ "$n" != "$1" ]; then
			# echo $n
			ip_address=$1
			host_name=$n
			# find existing instances in the host file and save the line numbers
			matches_in_hosts="$(grep -n "$host_name" /etc/hosts | cut -f1 -d:)"
			host_entry="${ip_address} ${host_name}"

			if [ -n "$matches_in_hosts" ]
			then
				# iterate over the line numbers on which matches were found
				for line_number in $matches_in_hosts; do
					# replace the text of each line with the desired host entry
					if [[ "$OSTYPE" == "darwin"* ]]; then
						$sudo_cmd sed -i '' "${line_number}s/.*/${host_entry} /" /etc/hosts
					else
						$sudo_cmd sed -i "${line_number}s/.*/${host_entry} /" /etc/hosts
					fi
				done
			else
				echo "$host_entry" | $sudo_cmd tee -a /etc/hosts > /dev/null
			fi
		fi
	done
}

set -e

domain="
static.guance.com
openway.guance.com
dflux-dial.guance.com
static.dataflux.cn
openway.dataflux.cn
dflux-dial.dataflux.cn
zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com
"

sudo_cmd=''
if type sudo >/dev/null 2>&1; then
	# detect root user
	if [ "$UID" != "0" ]; then
		sudo_cmd='sudo'
	fi
fi

##################
# colors
##################
RED="\033[31m"
CLR="\033[0m"

errorf() {
  msg=$1
  shift
  printf "${RED}[E] $msg ${CLR}\n" "$@" >&2
}

##################
# Set Variables
##################

# Detect OS/Arch

arch=
case $(uname -m) in

	"x86_64")
		arch="amd64"
		;;

	"i386" | "i686")
		arch="386"
		;;

	"aarch64")
		arch="arm64"
		;;

	"arm" | "armv7l")
		arch="arm"
		;;

	"arm64")
		arch="arm64"
		;;

	*)
		# shellcheck disable=SC2059
		printf "${RED}[E] Unsupported arch $(uname -m) ${CLR}\n"
		exit 1
		;;
esac

os="linux"

if [[ "$OSTYPE" == "darwin"* ]]; then
	if [[ $arch != "amd64" ]] && [[ $arch != "arm64" ]]; then # Darwin only support amd64 and arm64
		# shellcheck disable=SC2059
		printf "${RED}[E] Darwin only support amd64/arm64.${CLR}\n"
		exit 1;
	fi

	os="darwin"

	# NOTE: under darwin, for arm64 and amd64, both use amd64
	arch="amd64"
fi

printf "* Detect OS/Arch ${os}/${arch}\n"

# Select installer
installer_base_url="https://{{.InstallBaseURL}}"

if [ -n "$DK_INSTALLER_BASE_URL" ]; then
	installer_base_url=$DK_INSTALLER_BASE_URL
	printf "* Set installer_base_url => $DK_INSTALLER_BASE_URL\n"
fi

installer_file="installer-${os}-${arch}-{{.Version}}"
printf "* Detect installer ${installer_file}\n"

installer_url="${installer_base_url}/${installer_file}"
installer=/tmp/dk-installer

verbose_mode=
if [ -n "$DK_VERBOSE" ]; then
	verbose_mode="-v"
	printf "* Set verbose_mode => ON\n"
fi

dataway=
if [ -n "$DK_DATAWAY" ]; then
	dataway=$DK_DATAWAY
	printf "* Set dataway => $DK_DATAWAY\n"
fi

lite=
if [ -n "$DK_LITE" ]; then
	lite=$DK_LITE
	printf "* Set lite => ON\n"
fi

elinker=
if [ -n "$DK_ELINKER" ]; then
	elinker=$DK_ELINKER
	printf "* Set elinker => $DK_ELINKER\n"
fi

global_customer_keys=
if [ -n "$DK_SINKER_GLOBAL_CUSTOMER_KEYS" ]; then
	global_customer_keys=$DK_SINKER_GLOBAL_CUSTOMER_KEYS
	printf "* Set global_customer_keys => ${DK_SINKER_GLOBAL_CUSTOMER_KEYS}\n"
fi

dataway_sinker=
if [ -n "$DK_DATAWAY_ENABLE_SINKER" ]; then
	dataway_sinker=1
	printf "* Set dataway_sinker => ON\n"
fi

upgrade=
if [ -n "$DK_UPGRADE" ]; then
	upgrade=$DK_UPGRADE
	printf "* Set upgrade => ON\n"
fi

upgrade_manager=0
if [ -n "$DK_UPGRADE_MANAGER" ]; then
	upgrade_manager=$DK_UPGRADE_MANAGER
	printf "* Set upgrade_manager => ON\n"
fi

upgrade_ip_whitelist=
if [ -n "$DK_UPGRADE_IP_WHITELIST" ]; then
	upgrade_ip_whitelist=$DK_UPGRADE_IP_WHITELIST
	printf "* Set upgrade_ip_whitelist => ${DK_UPGRADE_IP_WHITELIST} \n"
fi

def_inputs=
if [ -n "$DK_DEF_INPUTS" ]; then
	def_inputs=$DK_DEF_INPUTS
	printf "* Set def_inputs => ${DK_DEF_INPUTS} \n"
fi

install_rum_symbol_tools=0
if [ -n "$DK_INSTALL_RUM_SYMBOL_TOOLS" ]; then
	install_rum_symbol_tools=1
	printf "* Set install_rum_symbol_tools => ON\n"
fi

http_public_apis=""
if [ -n "$DK_HTTP_PUBLIC_APIS" ]; then
	http_public_apis="$DK_HTTP_PUBLIC_APIS"
	printf "* Set http_public_apis => ${DK_HTTP_PUBLIC_APIS} \n"
fi

global_host_tags=
if [ -n "$DK_GLOBAL_HOST_TAGS" ]; then
	global_host_tags=$DK_GLOBAL_HOST_TAGS
	printf "* Set global_host_tags => ${DK_GLOBAL_HOST_TAGS} \n"
fi

global_election_tags=
if [ -n "$DK_GLOBAL_ELECTION_TAGS" ]; then
	global_election_tags=$DK_GLOBAL_ELECTION_TAGS
	printf "* Set global_election_tags => ${DK_GLOBAL_ELECTION_TAGS} \n"
fi

cloud_provider=
if [ -n "$DK_CLOUD_PROVIDER" ]; then
	cloud_provider=$DK_CLOUD_PROVIDER
	printf "* Set cloud_provider => ${DK_CLOUD_PROVIDER} \n"
fi

namespace=
if [ -n "$DK_NAMESPACE" ]; then
	namespace=$DK_NAMESPACE
	printf "* Set namespace => ${DK_NAMESPACE} \n"
fi

http_listen="localhost"
if [ -n "$DK_HTTP_LISTEN" ]; then
	http_listen=$DK_HTTP_LISTEN
	printf "* Set http_listen => ${DK_HTTP_LISTEN} \n"
fi

http_port=9529
if [ -n "$DK_HTTP_PORT" ]; then
	http_port=$DK_HTTP_PORT
	printf "* Set http_port => ${DK_HTTP_PORT} \n"
fi

install_only=0
if [ -n "$DK_INSTALL_ONLY" ]; then
	install_only=1
	printf "* Set install_only => ON \n"
fi

dca_white_list=""
if [ -n "$DK_DCA_WHITE_LIST" ]; then
	dca_white_list=$DK_DCA_WHITE_LIST
	printf "* Set dca_white_list => ${DK_DCA_WHITE_LIST} \n"
fi

dca_listen=""
if [ -n "$DK_DCA_LISTEN" ]; then
	dca_listen=$DK_DCA_LISTEN
	printf "* Set dca_listen => ${DK_DCA_LISTEN} \n"
fi

dca_enable=""
if [ -n "$DK_DCA_ENABLE" ]; then

	dca_enable="$DK_DCA_ENABLE"
	if [ -z "$dca_white_list" ]; then
		printf "${RED}[E] DCA service is enabled, but white list is not set in DK_DCA_WHITE_LIST!${CLR}\n"
		exit 1;
	fi

	printf "* Set dca_enable => ON \n"
fi

pprof_listen=""
if [ -n "$DK_PPROF_LISTEN" ]; then
	pprof_listen=$DK_PPROF_LISTEN
	printf "* Set pprof_listen => ${DK_PPROF_LISTEN} \n"
fi

ipdb_type=""
if [ -n "$DK_INSTALL_IPDB" ]; then
	ipdb_type=$DK_INSTALL_IPDB
	printf "* Set ipdb_type => ${DK_INSTALL_IPDB} \n"
fi

install_externals=""
if [ -n "$DK_INSTALL_EXTERNALS" ]; then
	install_externals=$DK_INSTALL_EXTERNALS
	printf "* Set install_externals => ON \n"
fi

if [ -n "$HTTP_PROXY" ]; then
	proxy=$HTTP_PROXY
	printf "* Set HTTP proxy => $HTTP_PROXY \n"
fi

if [ -n "$HTTPS_PROXY" ]; then
	proxy=$HTTPS_PROXY
	printf "* Set HTTPS proxy => $HTTPS_PROXY \n"
fi

# check nginx proxy
proxy_type=""
if [ -n "$DK_PROXY_TYPE" ]; then
	proxy_type=$DK_PROXY_TYPE
	proxy_type=$(echo "$proxy_type" | tr '[:upper:]' '[:lower:]') # => lowercase
	printf "* Set proxy type => $proxy_type\n"

	if [ "$proxy_type" = "nginx" ]; then
		# env DK_NGINX_IP has the highest priority on proxy level
		if [ -n "$DK_NGINX_IP" ]; then
			proxy=$DK_NGINX_IP
			if [ "$proxy" != "" ]; then
				printf "\n* Set nginx proxy => $DK_NGINX_IP \n"

				for i in $domain; do
					updateHosts "$proxy" "$i"
				done
			fi
			proxy=""
		fi
	fi
fi

env_hostname=
if [ -n "$DK_HOSTNAME" ]; then
	env_hostname=$DK_HOSTNAME
	printf "* Set env_hostname => $DK_HOSTNAME \n"
fi

limit_cpumax=30
if [ -n "$DK_LIMIT_CPUMAX" ]; then
	limit_cpumax=$DK_LIMIT_CPUMAX
	printf "* Set limit_cpumax => $DK_LIMIT_CPUMAX \n"
fi

limit_memmax=4096
if [ -n "$DK_LIMIT_MEMMAX" ]; then
	limit_memmax=$DK_LIMIT_MEMMAX
	printf "* Set limit_memmax => $DK_LIMIT_MEMMAX \n"
fi

limit_disabled=0
if [ -n "$DK_LIMIT_DISABLED" ]; then
	limit_disabled=1
	printf "* Set limit_disabled => ON \n"
fi

install_log=/var/log/datakit/install.log
if [ -n "$DK_INSTALL_LOG" ]; then
	install_log=$DK_INSTALL_LOG
	printf "* Set install_log => $DK_INSTALL_LOG \n"
fi

confd_backend=""
confd_basic_auth=""
confd_client_ca_keys=""
confd_client_cert=""
confd_client_key=""
confd_backend_nodes=""
confd_password=""
confd_scheme=""
confd_separator=""
confd_username=""
confd_access_key=""
confd_secret_key=""
confd_circle_interval=0
confd_confd_namespace=""
confd_pipeline_namespace=""
confd_region=""

if [ -n "$DK_CONFD_BACKEND" ]; then
	confd_backend=$DK_CONFD_BACKEND
fi

if [ -n "$DK_CONFD_BASIC_AUTH" ]; then
	confd_basic_auth=$DK_CONFD_BASIC_AUTH
fi

if [ -n "$DK_CONFD_CLIENT_CA_KEYS" ]; then
	confd_client_ca_keys=$DK_CONFD_CLIENT_CA_KEYS
fi

if [ -n "$DK_CONFD_CLIENT_CERT" ]; then
	confd_client_cert=$DK_CONFD_CLIENT_CERT
fi

if [ -n "$DK_CONFD_CLIENT_KEY" ]; then
	confd_client_key=$DK_CONFD_CLIENT_KEY
fi

if [ -n "$DK_CONFD_BACKEND_NODES" ]; then
	confd_backend_nodes=$DK_CONFD_BACKEND_NODES
fi

if [ -n "$DK_CONFD_PASSWORD" ]; then
	confd_password=$DK_CONFD_PASSWORD
fi

if [ -n "$DK_CONFD_SCHEME" ]; then
	confd_scheme=$DK_CONFD_SCHEME
fi

if [ -n "$DK_CONFD_SEPARATOR" ]; then
	confd_separator=$DK_CONFD_SEPARATOR
fi

if [ -n "$DK_CONFD_USERNAME" ]; then
	confd_username=$DK_CONFD_USERNAME
fi

if [ -n "$DK_CONFD_ACCESS_KEY" ]; then
	confd_role=$DK_CONFD_ACCESS_KEY
fi

if [ -n "$DK_CONFD_SECRET_KEY" ]; then
	confd_role=$DK_CONFD_SECRET_KEY
fi

if [ -n "$DK_CONFD_CIRCLE_INTERVAL" ]; then
	confd_role=$DK_CONFD_CIRCLE_INTERVAL
fi

if [ -n "$DK_CONFD_CONFD_NAMESPACE" ]; then
	confd_role=$DK_CONFD_CONFD_NAMESPACE
fi

if [ -n "$DK_CONFD_PIPELINE_NAMESPACE" ]; then
	confd_role=$DK_CONFD_PIPELINE_NAMESPACE
fi

if [ -n "$DK_CONFD_REGION" ]; then
	confd_role=$DK_CONFD_REGION
fi

git_url=""
if [ -n "$DK_GIT_URL" ]; then
	git_url=$DK_GIT_URL
	printf "* Set git_url => $DK_GIT_URL \n"
fi

git_key_path=""
if [ -n "$DK_GIT_KEY_PATH" ]; then
	git_key_path=$DK_GIT_KEY_PATH
	printf "* Set git_key_path => $DK_GIT_KEY_PATH \n"
fi

git_key_pw=""
if [ -n "$DK_GIT_KEY_PW" ]; then
	git_key_pw=$DK_GIT_KEY_PW
	printf "* Set git_key_pw => $DK_GIT_KEY_PW \n"
fi

git_branch=""
if [ -n "$DK_GIT_BRANCH" ]; then
	git_branch=$DK_GIT_BRANCH
	printf "* Set git_branch => $DK_GIT_BRANCH \n"
fi

git_pull_interval=""
if [ -n "$DK_GIT_INTERVAL" ]; then
	git_pull_interval=$DK_GIT_INTERVAL
	printf "* Set git_pull_interval => $DK_GIT_INTERVAL \n"
fi

enable_election=""
if [ -n "$DK_ENABLE_ELECTION" ]; then
	enable_election=$DK_ENABLE_ELECTION
	printf "* Set enable_election => $DK_ENABLE_ELECTION \n"
fi

rum_origin_ip_header=""
if [ -n "$DK_RUM_ORIGIN_IP_HEADER" ]; then
	rum_origin_ip_header=$DK_RUM_ORIGIN_IP_HEADER
	printf "* Set rum_origin_ip_header => $DK_RUM_ORIGIN_IP_HEADER \n"
fi

disable_404page=""
if [ -n "$DK_DISABLE_404PAGE" ]; then
	disable_404page=$DK_DISABLE_404PAGE
	printf "* Set disable_404page => $DK_DISABLE_404PAGE \n"
fi

log_level=""
if [ -n "$DK_LOG_LEVEL" ]; then
	log_level=$DK_LOG_LEVEL
	printf "* Set log_level => $DK_LOG_LEVEL \n"
fi

log=""
if [ -n "$DK_LOG" ]; then
	log=$DK_LOG
	printf "* Set log => $DK_LOG \n"
fi

gin_log=""
if [ -n "$DK_GIN_LOG" ]; then
	gin_log=$DK_GIN_LOG
	printf "* Set gin_log => $DK_GIN_LOG \n"
fi

user_name=""
if [ -n "$DK_USER_NAME" ]; then
	user_name=$DK_USER_NAME
	printf "* Set user_name => $DK_USER_NAME \n"
fi

crypto_aes_key=""
if [ -n "$DK_CRYPTO_AES_KEY" ]; then
	crypto_aes_key=$DK_CRYPTO_AES_KEY
	printf "* Set aes_key => $DK_CRYPTO_AES_KEY \n"
fi

crypto_aes_key_file=""
if [ -n "$DK_CRYPTO_AES_KEY_FILE" ]; then
	crypto_aes_key_file=$DK_CRYPTO_AES_KEY_FILE
	printf "* Set aes_key_file => $DK_CRYPTO_AES_KEY_FILE \n"
fi

printf "* Apply all DK_* envs done.\n"

##################
# Try install...
##################
# shellcheck disable=SC2059
printf "\n* Downloading installer ${installer} from ${installer_url}\n"

rm -rf $installer

if [ "$proxy" ]; then # add proxy for curl
	# shellcheck disable=SC2086
	curl $verbose_mode -x "$proxy" --fail --progress-bar $installer_url > $installer
else
	# shellcheck disable=SC2086
	curl $verbose_mode --fail --progress-bar $installer_url > $installer
fi

# Set executable
chmod +x $installer

if [ "$upgrade" ]; then
	# shellcheck disable=SC2059
	printf "\n* Upgrading DataKit...\n"
	$sudo_cmd $installer \
		--install-log="$install_log" \
		--upgrade --lite="${lite}" --elinker="${elinker}" --upgrade-manager="${upgrade_manager}" --upgrade-ip-whitelist="${upgrade_ip_whitelist}" --proxy="${proxy}" --installer_base_url="$installer_base_url"
else
printf "\n* Installing DataKit...\n"
$sudo_cmd $installer \
		--install-log="${install_log}" \
		--install-only="${install_only}" \
		--installer_base_url="${installer_base_url}" \
		--dataway="${dataway}" \
		--enable-inputs="${def_inputs}" \
		--http-public-apis="$http_public_apis" \
		--http-disabled-apis="$http_disabled_apis" \
		--install-rum-symbol-tools="$install_rum_symbol_tools" \
		--global-host-tags="${global_host_tags}" \
		--global-election-tags="${global_election_tags}" \
		--cloud-provider="${cloud_provider}" \
		--namespace="${namespace}" \
		--listen="${http_listen}" \
		--port="${http_port}" \
		--proxy="${proxy}" \
		--lite="${lite}" \
		--elinker="${elinker}" \
		--env_hostname="${env_hostname}" \
		--dca-enable="${dca_enable}" \
		--dca-listen="${dca_listen}" \
		--dca-white-list="${dca_white_list}" \
		--pprof-listen="${pprof_listen}" \
		--install-externals="${install_externals}" \
		--confd-backend="${confd_backend}" \
		--confd-basic-auth="${confd_basic_auth}" \
		--confd-client-ca-keys="${confd_client_ca_keys}" \
		--confd-client-cert="${confd_client_cert}" \
		--confd-client-key="${confd_client_key}" \
		--confd-backend-nodes="${confd_backend_nodes}" \
		--confd-password="${confd_password}" \
		--confd-scheme="${confd_scheme}" \
		--confd-separator="${confd_separator}" \
		--confd-username="${confd_username}" \
		--confd-access-key="${confd_access_key}" \
		--confd-secret-key="${confd_secret_key}" \
		--confd-circle-interval="${confd_circle_interval}" \
		--confd-confd-namespace="${confd_confd_namespace}" \
		--confd-pipeline-namespace="${confd_pipeline_namespace}" \
		--confd-region="${confd_region}" \
		--git-url="${git_url}" \
		--git-key-path="${git_key_path}" \
		--git-key-pw="${git_key_pw}" \
		--git-branch="${git_branch}" \
		--git-pull-interval="${git_pull_interval}" \
		--limit-cpumax="${limit_cpumax}" \
		--limit-memmax="${limit_memmax}" \
		--limit-disabled="${limit_disabled}" \
		--enable-election="${enable_election}" \
		--rum-origin-ip-header="${rum_origin_ip_header}" \
		--disable-404page="${disable_404page}" \
		--log-level="${log_level}" \
		--log="${log}" \
		--ipdb-type="${ipdb_type}" \
		--sinker-global-customer-keys="${global_customer_keys}" \
		--enable-dataway-sinker="${dataway_sinker}" \
		--user-name="${user_name}" \
		--crypto-aes_key="${crypto_aes_key}" \
		--crypto-aes_key_file="${crypto_aes_key_file}" \
		--upgrade-ip-whitelist="${upgrade_ip_whitelist}" \
		--gin-log="${gin_log}"
		fi
rm -rf $installer

# install completion
$sudo_cmd datakit tool --setup-completer-script
