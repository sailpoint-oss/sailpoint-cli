#!/bin/bash

#constants
VERSION="v2.4.2"
CHARON_MINIMUM_VERSION="1647"
ROOT_FS_MINIMUM_FREE_KB="2000000" #we want at least 2GB free normally
ROOT_FS_MINIMUM_FREE_KB_EMERGENCY="100000" # we must have at least 100 MB for things to function
CONFIG_YAML_FILE_PATH="/home/sailpoint/config.yaml"
MACHINE_ID="$(cat /etc/machine-id)"
IPADDR=$(networkctl status | grep Address | sed 's/Address: //' | grep -E -o '[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}')
DOCKER_PS_OUTPUT=$(sudo docker ps -s)
DOCKER_IMAGES_OUTPUT=$(sudo docker images)
STUNTOPTS=$1

# colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
INFO='\033[1;37m' #bold white
REDBOLDUL='\033[31;1;4m'
RESETCOLOR='\033[0m'

#### Pre-init Disk Check ####

# Special case for when VA is totally out disk space and this
# script is run via "curl | bash" because it can't be saved onto disk

check_enough_free_disk() {
  min_free_kb="$1"
  if [ -z "$min_free_kb" ]; then
    echo "false"
    return 1
  fi
  root_free_kb=$(df -k | grep " /$" | awk '{print $4}')
  echo "Root volume nas $root_free_kb kb free". 2>&1
  if [ "$root_free_kb" -gt "$ROOT_FS_MINIMUM_FREE_KB_EMERGENCY" ]; then
    echo "true"
    return 0
  else
    echo "false"
    return 1
  fi
}

#check if we're being run interactively or not

if [ -t 0 ]; then
  echo "Welcome to SailPoint's VA Troubleshooting Script"
else
  echo "Welcome to SailPoint's VA Troubleshooting Script"
  echo "Script might be executed from curl | bash because there is not enough free disk space"
  # check free space

  if check_enough_free_disk $ROOT_FS_MINIMUM_FREE_KB_EMERGENCY ; then
    echo "This VA has at least $ROOT_FS_MINIMUM_FREE_KB_EMERGENCY kb root disk free"
  else
    echo "This VA is critically low on disk space. Attempting cleanup"
    current_images=$( echo "$DOCKER_IMAGES_OUTPUT" | grep 'current' | awk '{print "-e " $3}' | tr "\n" " ")
    echo "$DOCKER_IMAGES_OUTPUT" | grep -v $current_images -e REPOSITORY | awk '{print $1 ":" $2}' | xargs sudo docker rmi
    sudo rm -f /home/sailpoint/log/*.{0,1}  # delete rotated logs
    sudo journalctl --no-pager --rotate
    sudo journalctl --no-pager --vacuum-time=1d
    # checking if that was enough space freed
    if check_enough_free_disk $ROOT_FS_MINIMUM_FREE_KB_EMERGENCY; then
      echo "Freed sufficient space to continue"
    else
      # still not enough, truncate logs
      echo "Still not enough free space, truncating log files"
      sudo chown -R sailpoint /home/sailpoint/log #make sure we own the files
      echo "Stoping all containers to truncate logs. Will reboot after to restart containers"
      sudo docker ps -q | xargs sudo docker stop
      truncate -s 0 /home/sailpoint/log/fluent.log /home/sailpoint/log/ccg.log
      sync
      sleep 1
      echo "Log files truncated. New free space is $(df -k | grep " /$" | awk '{print $4}')kb. Rebooting"
      sleep 5 #giving time for user to read echo
      sudo reboot
    fi
  fi
fi

### INIT ###
if [[ $(sudo systemctl status canal | grep enabled | wc -l) -gt 0 ]]; then
  IS_CANAL_ENABLED=true
else
  IS_CANAL_ENABLED=false
fi

if [[ -e "$CONFIG_YAML_FILE_PATH" ]]; then
  # Set global vars whose data come from config.yaml
  CONFIG_YAML=$(< $CONFIG_YAML_FILE_PATH )
  ORGNAME=$( echo "$CONFIG_YAML" | grep -oP '(?<=org: ).*' )
  ORGNAME="${ORGNAME//$'\r'/}"                                #remove return characters
  PODNAME=$( echo "$CONFIG_YAML" | grep -oP '(?<=pod: ).*' )
  PODNAME="${PODNAME//$'\r'/}"                                #remove return characters
  # detect Canal in systemd
  #https://app.asana.com/1/40019095804142/project/1212671613187404/task/1212991215552901?focus=true
else
  echo "*** Config file not found. "
  echo "*** Would you like to create a temporary config.yaml so stunt can run?"
  echo "*** This file can be overwritten manually, or removed by resetting the"
  echo "*** VA with 'va-bootstrap reset'."
  read -p "[y\\n] > " response
  case $response in
    [Yy])
      echo "Generating file..."
      ORGNAME="mytestorg"
      PODNAME="stg01-useast1"
      touch $CONFIG_YAML_FILE_PATH && 
      echo -e "pod: $PODNAME\norg: $ORGNAME\napiUser: \"testapiuser\"\napiKey: \"::!:A2:testapikey\"\nkeyPassphrase: \"::!:A2:testkeypassphrase\"" > $CONFIG_YAML_FILE_PATH
      if [-f $CONFIG_YAML_FILE_PATH ]; then
        echo "File generated successfully."
      else 
        echo "ERROR: Unknown error when checking for existence of test config.yaml; exiting." 
        endscript
      fi
    ;;
    *)
      echo 
    ;;
  esac
fi

### GLOBAL RUNTIME VARIABLES ###

DATE=$(date -u +"%b_%d_%y-%H_%M")
DIVIDER="================================================================================"
ZIPFILE=/home/sailpoint/logs.$ORGNAME-$PODNAME-$(hostname)-$IPADDR-$DATE.zip # POD-ORG-CLUSTER_ID-VA_ID.zip. 
LISTOFLOGS="/home/sailpoint/log/*.log"
CCGDIR="/home/sailpoint/ccg/"
RUNNING_FLATCAR_VERSION="$(cat /etc/os-release | grep -oP 'VERSION=\K[^<]*')"
FLATCAR_RELEASES_URL="https://www.flatcar.org/releases"
FLATCAR_STABLE_RELEASE_FILE="https://stable.release.flatcar-linux.net/amd64-usr/current/version.txt"
CERT_DIRECTORY="/home/sailpoint/certificates"
ADD_REBOOT_MESSAGE=false # Flip if reboot is required; makes a colorful message appear on stdout
PROXY_FILE_PATH="/home/sailpoint/proxy.yaml"
AWS_REGIONS=("us-east-1", "us-west-2", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ca-central-1", "eu-central-1", "eu-west-2")
AWS_REGION="us-east-1"
ISC_DOMAIN="identitynow.com"
ISC_ACCESS="accessiq.sailpoint.com"
JAVA_OVERWRITES_FILE_PATH="/home/sailpoint/ccg/java_overwrites.yaml"
IS_CCG_RUNNING=false
IS_IAI_VA=$(echo "$CONFIG_YAML" | grep -iq "iai:" && echo true || echo false)
LOGFILE=/home/sailpoint/stuntlog-$ORGNAME-$IPADDR.log
gather_logs=true

# Get main partition name - CS0334359
if [[ $(findmnt -nro SOURCE / ) ]]; then
  MAIN_PARTITION=$(findmnt -nro SOURCE /)
else
  echo "Error finding root mount point. Defaulting to static string \"/sda9|nvme0n1p9/\" for testing."
  MAIN_PARTITION="/sda9|nvme0n1p9/"
fi

# FedRAMP compatibility - CS0305183
FEDRAMP_STRING="usgov"
IS_ORG_FEDRAMP=false
AWS_FEDRAMP_REGIONS=("us-gov-west-1", "us-gov-east-1")
# Check the pod string for "usgov"; if found, change bool 
# Create URLs like this: "https://$ORGNAME.$ISC_DOMAIN", "https://sqs.$REGION.amazonaws.com", "https://$ORGNAME.$ISC_ACCESS"
if [[ $PODNAME == *"$FEDRAMP_STRING"* ]]; then
  IS_ORG_FEDRAMP=true
  AWS_REGION="us-gov-west-1"
  ISC_DOMAIN="saas.sailpointfedramp.com"
  ISC_ACCESS="idn.sailpointfedramp.com"
fi

# identitynow-demo.com compatibility - CS0390503
IS_ORG_DEMO=false
DEMO_STRING1="-poc"
DEMO_STRING2="partner"
DEMO_STRING3="training"
DEMO_STRING4="company"
if [[ ($PODNAME == *"$DEMO_STRING1"*) || ($PODNAME == *"$DEMO_STRING2"*) || ($PODNAME == *"$DEMO_STRING3"*) || ($PODNAME == *"$DEMO_STRING4"*) ]]; then
  IS_ORG_DEMO=true
  AWS_REGION="us-east-1"
  ISC_DOMAIN="identitynow-demo.com"
  ISC_ACCESS="accessiq.identitynow-demo.com"
fi

# for test pass/warn/fail and summary
total_tests=0
passes=0
failures=0
warnings=0
declare -A test_categories
summary=""
output_document_summary="$summary"

# for error handling
error_message=""

# for curl test
total_seconds_to_test=180
seconds_between_tests=4

### FUNCTIONS ###

help () {
  # Display help
  echo "Stunt version: $VERSION"
  echo "The stunt script collects information about the current state of your"
  echo "VA, and places that data into a stuntlog text file in your home directory."
  echo "Collecting this helps SailPoint Support Engineers troubleshoot your system."
  echo
  echo "Syntax: ./stunt.sh [-h] [ -t,p,f,o,s,j,l/L  OR  n|u|c ]"
  echo "Options:"
  echo "h   Print this help info then exit"
  echo "t   Add traceroute test to SQS"
  echo "p   Add ping test"
  echo "f   Add automatic fixup steps"
  echo "o   Add openssl config file into stuntlog"
  echo "s   Pulls a server-based certificate from a server and installs it in /home/sailpoint/certificates/"
  echo "l/L Add zip collection of log files and archive them along with stuntlog file (Default)"
  echo "j   Add to zip collection of the last day of the systemd journal"
  echo "n   [exclusive] Do not collect archive of log files; this overrides zip file collection"
  echo "u   [exclusive] Only perform forced OS update (this will make system changes) then exit"
  echo "c   [exclusive] Only perform a curl test that connects to SQS and S3, one test every four seconds for three minutes then exit"
}

# Get cmd line args
while getopts ":htpfoslLjnucr" option; do
  case $option in
    h) #display help
      help
      exit;;
    t)
      do_traceroute=true;;
    p)
      do_ping=true;;
    f)
      do_fixup=true;;
    o)
      get_ssl_conf=true;;
    s)
      get_iqs_cert_process=true;;
    l)
      gather_logs=true;;
    L)
      gather_logs=true;;
    j)
      capture_journal=true;;
    n)
      gather_logs=false;;
    u)
      do_update=true;;
    c)
      curl_test=true;;
    \?)
      echo "Invalid argument on command line. Please review help below:"
      help
      exit;;
    esac
done

# args:
# $1 == stdout description
intro() {
  set -f #Disable globbing
  echo "$DIVIDER" >> "$LOGFILE"
  echo "$1"
  echo "$1" >> "$LOGFILE"
  echo "$DIVIDER" >> "$LOGFILE"
  set +f
}

outro() {
  set -f
  echo >> "$LOGFILE"
  echo
  set +f
}

expect() {
  set -f
  echo "********************************************************************************" >> "$LOGFILE"
  echo "*** Expect $1" >> "$LOGFILE"
  echo "********************************************************************************" >> "$LOGFILE"
  echo >> "$LOGFILE"
  set +f
}

endscript() {
  echo "$DIVIDER"
  intro "$(date -u) - END TESTS for $ORGNAME on $PODNAME "
  echo >&2 "*** Tests completed on $(date -u) ***"
  echo "$DIVIDER"
  if [ "$ADD_REBOOT_MESSAGE" == true ]; then
    echo -e "$YELLOW$DIVIDER $REDBOLDUL"
    echo -e "REBOOT IS REQUIRED; PLEASE RUN 'sudo reboot' NOW $RESETCOLOR"
    echo -e "$YELLOW$DIVIDER $RESETCOLOR"
  fi
}

get_keyPassphrase_length() {
  cat $CONFIG_YAML_FILE_PATH | grep -E "keyPassphrase:\s*([\"'])([^\"']*)\1" | sed -E "s/keyPassphrase: ['\"]//g" | sed -E "s/['\"]$//gm" | wc -m # Will return 0 if unencrypted
}

get_num_share_jobs() {
  echo $(find "/opt/sailpoint/share/jobs" -maxdepth 1 -type f | wc -l)
}
get_num_workflow_jobs() {
  echo $(find "/opt/sailpoint/workflow/jobs" -maxdepth 1 -type f | wc -l)
}

add_test_result() {
  local category="$1"
  local test_name="$2"
  local test_result="$3"
  local test_output="$4"

  if [[ -z "${test_categories[$category]}" ]]; then
    test_categories["$category"]="$test_name: $test_result -- Test output: $test_output"
  else
    test_categories["$category"]="${test_categories["$category"]},$test_name: $test_result -- Test output: $test_output"
  fi
}

# CS0237804
# example:
# perform_test test_name test_command pass_comparison_operator pass_expected_condition fail_comparison_operator fail_expected_condition test_category
# e.g.: perform_test "Does 1 = 1?" "if [[ 1 == 1 ]]; then echo true; fi" "==" "true" "==" "false" "<null>" "system"
perform_test() {
  local test_name="$1"
  local test_command="$2"
  local pass_comparison_operator="$3"
  local pass_expected_condition="$4"
  local fail_comparison_operator="$5"
  local fail_expected_condition="$6"
  local test_category="$7"

  output=$(eval "$test_command")

  echo -e $DIVIDER

  if [ "$pass_comparison_operator" = "==" ] && [ "$output" = "$pass_expected_condition" ]; then
    echo -e "PASS: $test_name" >> "$LOGFILE"
    echo -e "${GREEN}PASS$RESETCOLOR: $test_name"
    ((passes++))
    add_test_result "$test_category" "$test_name" "pass" "$output"
  elif [ "$fail_comparison_operator" = "==" ] && [ "$output" = "$fail_expected_condition" ]; then
    echo -e "FAIL: $test_name" >> "$LOGFILE"
    echo -e "${RED}FAIL$RESETCOLOR: $test_name"
    ((failures++))
    add_test_result "$test_category" "$test_name" "fail" "$output"
  elif [ "$pass_comparison_operator" != "==" ] && [ "$output" "$pass_comparison_operator" "$pass_expected_condition" ]; then
    echo -e "PASS: $test_name" >> "$LOGFILE"
    echo -e "${GREEN}PASS$RESETCOLOR: $test_name"
    ((passes++))
    add_test_result "$test_category" "$test_name" "pass" "$output"
  elif [ "$fail_comparison_operator" != "==" ] && [ "$output" "$fail_comparison_operator" "$fail_expected_condition" ]; then
    echo -e "FAIL: $test_name" >> "$LOGFILE"
    echo -e "${RED}FAIL$RESETCOLOR: $test_name"
    ((failures++))
    add_test_result "$test_category" "$test_name" "fail" "$output"
  else
    echo -e "WARNING: $test_name" >> "$LOGFILE"
    echo -e "${YELLOW}WARNING$RESETCOLOR: $test_name"
    ((warnings++))
    add_test_result "$test_category" "$test_name" "warn" "$output"
  fi

  echo -e $DIVIDER
}

output_all_tests_by_category() {
  for category in "${!test_categories[@]}"; do
    echo "Category: $category"
    tests="${test_categories[$category]}"
    IFS=',' read -ra test_results <<< "$tests"

    for test_result in "${test_results[@]}"; do
      echo "  $test_result"
    done
  done
}

# Handle exceptions
handle_interrupts() {
  intro "Script interrupted by ctrl+c. Exiting."
  echo "The stuntlog file at $LOGFILE is incomplete due to the above exception."
  endscript
  exit 0
}

handle_error() {
  error_message="$1"
  intro "Script encountered an error with the following command: ${BASH_COMMAND}."
  echo "The stuntlog file at $LOGFILE contains an error due to the exception. Continuing... "
  echo -e "Test $test_name: WARNING" >> "$LOGFILE"
  echo -e "Test $test_name:${YELLOW} WARNING $RESETCOLOR"
  add_test_result "script" "Error handler" "warn" "$1"
  ((warnings++))
  outro
}

trap handle_interrupts SIGINT
trap 'handle_error "$BASH_COMMAND"' ERR

is_ip_private() {
  local private_ranges=(
    "10.0.0.0/8"
    "172.16.0.0/12"
    "192.168.0.0/16"
  )

  IFS='.' read -r -a ip_parts <<< "$IPADDR"
  local ip_number=$(( (${ip_parts[0]} * 256**3) + (${ip_parts[1]} * 256**2) + (${ip_parts[2]} * 256) + ${ip_parts[3]} )) 

  for range in "${private_ranges[@]}"; do
    IFS='/' read -r -a range_parts <<< "$range"
    local network_address="${range_parts[0]}"
    local subnet_mask="${range_parts[1]}"

    IFS='.' read -r -a network_parts <<< "$network_address"
    local range_number=$(( (${network_parts[0]} * 256**3) + (${network_parts[1]} * 256**2) + (${network_parts[2]} * 256) + ${network_parts[3]} ))
    local mask=$((0xFFFFFFFF * (2 ** (32 - subnet_mask)) ))

    if (( (ip_number & mask) == (range_number & mask) )); then
      echo 0
      return
    fi
  done

  echo 1
}

cert_tester() {
  failed_certs=()
  starting_failures=$failures

  # Check if the directory exists
  if [ ! -d "$CERT_DIRECTORY" ]; then
    echo "Directory $CERT_DIRECTORY does not exist."
    intro "ERROR: $CERT_DIRECTORY was not found; skipping test."
    add_test_result "configuration" "Does $CERT_DIRECTORY exist?" "warn" "$CERT_DIRECTORY not found"
    ((warnings++))
    return
  fi

  # Loop through each certificate file in the directory
  for cert_file in "$CERT_DIRECTORY"/*; do
    expired_certs=()
    if [ -f "$cert_file" ]; then
      # Check if the file is in PEM format using the 'openssl' command
      if openssl x509 -in "$cert_file" -noout -text &>/dev/null; then
        # Certificate is valid (in PEM format)
        add_test_result "certificates" "Cert test: $cert_file" "pass" "Certificates"
        ((passes++))
        if ! openssl x509 -checkend 0 -noout -in "$cert_file"; then
          expired_certs+=("$cert_file")
        fi
      else
        # Certificate is not in PEM format
        add_test_result "certificates" "Cert test: $cert_file" "fail" "Certificates"
        failed_certs+=("$(basename "$cert_file")")
        ((failures++))
      fi
    else
      echo "Skipped: $cert_file (Unknown file type)"
      add_test_result "certificates" "Cert test: $cert_file" "warn" "Certificates"
      failed_certs+=("$(basename "$cert_file")")
      ((warnings++))
    fi
  done

  if [[ $failures > $starting_failures ]]; then
    echo "Failed or skipped certificates:"
    for cert_name in "${failed_certs[@]}"; do
      echo -e "Test - $RED FAIL$RESETCOLOR: $cert_name"
    done
  fi

  if [ ! ${#expired_certs[@]} -eq 0 ]; then
    echo "Expired certificates: " | tee -a "$LOGFILE"
    for i in "${!expired_certs[@]}"; do
      echo "$((i + 1)). ${expired_certs[i]}"
    done
    if [[ "$do_fixup" == true ]]; then
      read -p "Delete expired certs? (Y/N): " choice
      if [[ "$choice" =~ ^[Yy]$ ]]; then
        echo "Choice read in: $choice" | tee -a "$LOGFILE"
        for cert in "${expired_certs[@]}"; do
          rm "$cert"
          echo "Deleted $cert" | tee -a "$LOGFILE"
        done
      else
        echo "No certificates deleted." | tee -a "$LOGFILE"  
      fi
    else
      echo
      echo -e "${YELLOW}WARNING:$RESETCOLOR Expired/expiring certificates found, but not deleted." | tee -a "$LOGFILE"
      echo -e "${YELLOW}ACTION:$RESETCOLOR Re-run stunt with -f to delete expired certificate files." | tee -a "$LOGFILE"
      echo
    fi
  fi
}

# Gives integer expression expected
canalenv_exists() {
  if [[ -e /home/sailpoint/canal.env ]]; then
    echo 0
  else
    if [[ "$do_fixup" == true ]]; then
      echo "Creating canal.env"
      touch /home/sailpoint/canal.env
    else
      echo -e "${YELLOW}ACTION:$RESETCOLOR File does not exist, but the option for automatic fixup "
      echo -e "is not enabled. Rerun STUNT with -f to create the canal.env file"
    fi
    echo 1
  fi
}

canal_log_contains_FNF_string() {
  if [[ $(cat /home/sailpoint/log/canal-hc.log | grep "No such file or directory" | wc -l) -lt 1 ]]; then
    echo 0
  else
    if [[ "$do_fixup" == true ]]; then
      echo "Creating canal-hc.log" | tee -a "$LOGFILE"
      touch /home/sailpoint/log/canal-hc.log | tee -a "$LOGFILE"
    else
      echo -e "${YELLOW}ACTION: $RESETCOLOR 'No such file or directory' error found in canal-hc.log," | tee -a "$LOGFILE"
      echo -e "but the option for automatic fixup is not enabled. Rerun STUNT with -f to create the canal-hc.log file" | tee -a "$LOGFILE"
    fi
    echo 1
  fi
}

canal-hc_log_exists() {
  if [[ -e /home/sailpoint/log/canal-hc.log ]]; then
    echo 0
  else
    echo 1
  fi
}

ntp_sync() {
  if timedatectl show --property=NTPSynchronized --value | grep -q '^yes$'; then
    echo 0
  else
    echo 1
  fi
}

get_charon_network_test_line() {
  grep -a 'Networking check' "/home/sailpoint/log/charon.log" | tail -1
}

detect_old_os_version() {
  #look in $RUNNING_FLATCAR_VERSION for major version 2345 or lower
  major_version=$( echo "$RUNNING_FLATCAR_VERSION" | awk -F'.' '{ print $1 }' )
  if [[ $major_version -le 2345 ]]; then
    echo "Major version is $major_version, and requires update." | tee -a "$LOGFILE"
    if [[ "$do_fixup" == true ]]; then
      echo -e "${INFO}INFO$RESETCOLOR: Since -f flag was used, we'll attempt to update automatically." | tee -a "$LOGFILE"
      echo | tee -a "$LOGFILE"
      update_old_os
      echo 0
    else
      echo -e "${YELLOW}ACTION:$RESETCOLOR Old OS detected, but fixup flag is not enabled." | tee -a "$LOGFILE"
      echo -e "Rerun STUNT with -f to attempt the automatic update process." | tee -a "$LOGFILE"
      echo 1
    fi
  else # major version greater than 2345
    echo 0
  fi
}

get_flatcar_current_version() { #"https://www.flatcar.org/releases"
  if flatcar_html=$(curl -s -L --connect-timeout $seconds_between_tests $FLATCAR_STABLE_RELEASE_FILE 2>/dev/null); then
    FLATCAR_CURRENT_VER=$(curl -fsSL $FLATCAR_STABLE_RELEASE_FILE | grep FLATCAR_VERSION= | cut -d = -f 2)
    echo $FLATCAR_CURRENT_VER
  else
    echo "Error: unable to fetch HTML content from $FLATCAR_STABLE_RELEASE_FILE" >> "$LOGFILE"
  fi
}

get_update_engine_status() {
  timeout 2 sudo update_engine_client -status
}

no_proxy_double_quotes() {
  no_proxy_value=$(grep "^no_proxy:" "$PROXY_FILE_PATH" | awk -F': ' '{print $2}')

  if [[ "$no_proxy_value" =~ ^\".*\"$ ]]; then
    echo 0
  else
    echo 1
  fi
}

get_current_image_tag() {
  image_name="$1"
  current_image_id=$(echo "$DOCKER_IMAGES_OUTPUT" | grep "$image_name" | grep current | head -n 1 | awk '{print $3}')
  current_image_tag=$(echo "$DOCKER_IMAGES_OUTPUT" | grep "$image_name" | grep "$current_image_id" | grep -v current | awk '{print $2}' | head -n 1)
  echo "$current_image_tag" | grep -o '^[[:digit:]]*'
}

clean_non_current_images() {
  echo "Cleaning images, errors can be ignored"
  current_images=$(echo "$DOCKER_IMAGES_OUTPUT" | grep 'current' | awk '{print "-e " $3}' | tr "\n" " ")
  echo "$DOCKER_IMAGES_OUTPUT" | grep -v $current_images -e REPOSITORY | awk '{print $1 ":" $2}' | xargs sudo docker rmi
  echo "Cleaning is complete"
}

fix_missing_images() {
  if [ -f "/opt/sailpoint/share/ecs/ecs-auth.json" ]; then
    ecr_creds=$(cat /opt/sailpoint/share/ecs/ecs-auth.json)
    ecr_pw=$(jq -r '.pwd' < /opt/sailpoint/share/ecs/ecs-auth.json)
    ecr_host=$(jq -r '.repo' < /opt/sailpoint/share/ecs/ecs-auth.json | sed 's|https://||' )
    echo "$ecr_pw" | sudo docker login --username AWS --password-stdin ${ecr_host}/sailpoint/charon
    sudo systemctl stop charon
    sudo docker pull ${ecr_host}/sailpoint/charon && \
      sudo docker tag ${ecr_host}/sailpoint/charon:latest ${ecr_host}/sailpoint/charon:current && \
      sudo systemctl start charon && \
      sleep 10
    charon_status=$(systemctl list-units charon.service --no-pager --output json | jq -r '.[0].sub')
    ADD_REBOOT_MESSAGE=true
    if [ "$charon_status" = "running" ]; then
      echo "Succesfully pulled updated Charon image" | tee -a "$LOGFILE"
    else
      echo "ERROR: Failed to pull and start latest Charon. Please contact Support" | tee -a "$LOGFILE"
    fi
    return 0

  else
    echo "ERROR: No ECR Creds on disk, can not attempt missing image repair. Please contact Support with this error." >> "$LOGFILE"
  fi
}

test_openssh_version () {
  required_major_version=10
  required_minor_version=0 # require v10.0 or above
  openssh_version_output=$(ssh -V 2>&1)
  openssh_version=$(echo "$openssh_version_output" | grep -oP '(?<=OpenSSH_)[0-9]+\.[0-9]+')
  openssh_major_version=$(echo $openssh_version | cut -d '.' -f 1)
  openssh_minor_version=$(echo $openssh_version | cut -d '.' -f 2)

  if (( $openssh_major_version >= $required_major_version )); then
    # Current version is updated
    if (( $openssh_minor_version >= $required_minor_version )); then
      echo 0
    fi
  else
    echo 1
  fi
}

check_container_running () {
  if [[ $( echo "$DOCKER_PS_OUTPUT" | grep $1 | wc -l) -gt 0 ]]; then 
    echo true; 
  else 
    echo false; 
  fi
}

canal_connection_test () {
  echo -e '\x00\x0e\x38\xa3\xcf\xa4\x6b\x74\xf3\x12\x8a\x00\x00\x00\x00\x00' | ncat $1 443 | head -c 5 | cat -v | tr -d '[:space:]' | grep -e @^Z@ | wc -m;
}

#test string:
# "100.80.9.9|localhost|*.googleapis.com|*.db.com|*.internal|*.mybad.io|169.254.169.254|10.216.204.118|127.0.0.1|100.80.99.244"
check_no_proxy_validate_format() {
  local value=$(awk -F': ' '/^no_proxy:/ {print $2}' "$PROXY_FILE_PATH")
  if [[ ! "$value" =~ ^\"?[a-zA-Z0-9|*.]+\"?$ ]]; then
    echo 1
    return
  fi
  
  # Success:
  echo 0
}

get_lscpu_num_cpus () {
  lscpu | grep -i ^CPU\(s | awk '{print $2}'
}

get_free_total_mem_in_gb () {
  free -h | grep "Mem:" | awk -F' ' '{print $2}' | sed 's/Gi//' | xargs printf "%.0f" 
}

get_network_adapter_name () {
  ip -o link show | awk -F': ' '/state UP/ && $2 != "lo" {print $2; exit}'
}

get_iqservice_cert () {
  local iqservice_network_address
  local iqservice_secure_port
  local repeat=true
  
  while $repeat; do
    read -p "Enter the IP address or hostname of the IQService server: " iqservice_network_address
    read -p "Enter the TLS port for the IQService (usually 5050 or 5051): " iqservice_secure_port
    local cert_filepath="/home/sailpoint/certificates/${iqservice_network_address}.cer"
    timeout 7 openssl s_client -connect $iqservice_network_address:$iqservice_secure_port >&2 | grep -Pzo '(?s)-----BEGIN CERTIFICATE-----.*-----END CERTIFICATE-----' > $cert_filepath;
    sync

    if test -f $cert_filepath ; then                                                                          # if new cert exists,
      if [ $(stat -c%s $cert_filepath) -gt 0 ]; then                                                          # and it isn't empty
        echo -e "${GREEN}SUCCESS ${RESETCOLOR}- Certificate was added at $cert_filepath" | tee -a "$LOGFILE"  # then success
        sudo systemctl restart ccg
        echo
      else
        echo -e "${YELLOW}WARN ${RESETCOLOR}- New certificate file is blank; check your server string and port information, then try again." | tee -a "$LOGFILE"
        echo -e "- This is usually due to a timeout with no immediate response from the server. Errors may or may not be shown."
      fi
    else
      echo -e "${RED}FAIL ${RESETCOLOR}- The file for ${cert_filepath} was not found - check your server string and port information, then try again." | tee -a "$LOGFILE"
      echo -e "-"
      endscript
      exit 1
    fi
    read -p "Pull another cert? [y/n]: " again
    case $again in
      [Yy])
        repeat=true
      ;;
      *)
        repeat=false
        echo -e "Ending script"
      ;;
    esac
  done
  endscript
  echo "EXITING"
  exit 0
}

get_top_output() {
  top -b -n 1 | awk '
  BEGIN {FS=" "}
  NR==1 {print; next}
  NR==2 {print; next}
  NR==3 {print; next}
  NR==4 {print; next}
  NR==5 {print; next}
  NR==6 {print; next}
  NR==7 {print; next}
  /sailpoi+/ {print $0}
  /docker/ {print $0}
'
}


### END FUNCTIONS ###


### START STDOUT OUTPUT ###

echo $DIVIDER
echo "STUNT -- Support Team UNified Test -- ${VERSION}"
echo $DIVIDER
echo "*** This script tests network connectivity, gathers log/system data,"
echo "*** performs recommended setup steps from the SailPoint VA documents which"
echo "*** when skipped will cause network connectivity problems, and creates a"
echo "*** log file at '$LOGFILE'."
echo "*** No warranty is expressed or implied for this tool by SailPoint."
echo


if test -f "$LOGFILE"; then
  echo "*** Found an old log file. Renaming..." &&
  mv $LOGFILE $LOGFILE.$DATE.old
else
  touch $LOGFILE
fi

virt_host=$(systemd-detect-virt)

# Start the tests by placing a header in the logfile
echo $DIVIDER | tee -a "$LOGFILE"
echo "$(date -u) - STARTING TESTS for $ORGNAME on $PODNAME"
echo $DIVIDER
echo "*** STARTING TESTS ***" | tee -a "$LOGFILE"
echo "Date:                 $(date -u)" | tee -a "$LOGFILE"
echo "Stunt ver.:           $VERSION" | tee -a "$LOGFILE"
echo "Org:                  $ORGNAME" | tee -a "$LOGFILE"
echo "Pod:                  $PODNAME" | tee -a "$LOGFILE"
echo "IP Address:           $IPADDR" | tee -a "$LOGFILE"
echo "Machine-id:           $MACHINE_ID" | tee -a "$LOGFILE"
echo "Canal enabled:        $IS_CANAL_ENABLED" | tee -a "$LOGFILE"
echo "Virtualization host:  $virt_host" | tee -a "$LOGFILE"
echo "Flags used for stunt: $STUNTOPTS" | tee -a "$LOGFILE"
echo $DIVIDER >> "$LOGFILE"
echo "<SUMMARY_BLOCK>" >> "$LOGFILE"
echo $DIVIDER >> "$LOGFILE"
outro

# Download TLS cert from IQService process
if [[ $get_iqs_cert_process == true ]]; then
  get_iqservice_cert
fi

# Forced update process

## Forced update functions
update_old_OS_with_new_charon() {
  echo "Updating OS with 'flatcar-update -Q'. This will require a reboot of the VA when complete."
  version=$(get_flatcar_current_version)
  if [[ -e /etc/systemd/system/update-engine.service.d/override.conf ]]; then
    sudo rm /etc/systemd/system/update-engine.service.d/override.conf
  fi
  if [[ $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    sudo /opt/sailpoint/share/bin/flatcar-update -Q --to-version $version
  else 
    echo -e "${YELLOW}WARNING:$RESETCOLOR Unable to gather version information from flatcar website; trying with default OS version value of 4459.2.0."
    sudo /opt/sailpoint/share/bin/flatcar-update -Q --to-version 4459.2.0
  fi
  sudo rm /etc/systemd/system/update-engine.service.d/override.conf
  echo 0;
}

reset_machine_id() {
  local MACHINE_ID="$(cat /etc/machine-id)"
  echo "Old ID: $MACHINE_ID"
  sudo rm -f /etc/machine-id  >> "$LOGFILE" 2>&1
  sudo systemd-machine-id-setup  >> "$LOGFILE" 2>&1
  local new_machine_id=$(cat /etc/machine-id)
  echo "New ID: $new_machine_id"
  if [[ $MACHINE_ID == $new_machine_id ]]; then
    echo "Old and new IDs match; interrupting update process and forcing a reboot" | tee -a "$LOGFILE"
    sleep 5
    endscript
    sudo rm -f /etc/machine-id && sudo reboot
  fi
}

update_old_os() {
  current_charon=$(get_current_image_tag charon)
  if [[ $current_charon -gt 2239 ]]; then
    intro "Attempting modern update because of newer version of charon: v$current_charon. This will require a reboot of the VA when successful."
    echo "Current charon version: $current_charon"
    update_old_OS_with_new_charon
  else
    intro "Older version of charon: $current_charon. Retrieving Flatcar Linux certificate, and starting OS update. This will require a reboot of the VA when successful."
    echo "This will reboot the VA when complete."
    curl -sS https://stable.release.flatcar-linux.net/ >/dev/null
    sudo rm /etc/ssl/certs/DST_Root_CA_X3.pem
    sudo update-ca-certificates
    curl -sS https://stable.release.flatcar-linux.net/ >/dev/null

    # check for existence of sudo permissions to run systemctl restart update-engine; if exists, run it, else try old method
    if [[ $(sudo -l | grep "systemctl restart update-engine" | wc -l) -gt 0 ]]; then
      update_engine_result=$(sudo systemctl restart update-engine 2>&1)
      echo "$update_engine_result" | tee -a "$LOGFILE"
    else
      sudo update_engine_client -reset_status && sudo update_engine_client -update >> "$LOGFILE" 2>&1
    fi
  fi
  if [[ $(grep "UPDATE_STATUS_UPDATED_NEED_REBOOT" $LOGFILE | wc -l) -gt 0  ]]; then
    ADD_REBOOT_MESSAGE=true
  fi
}

if [ "$do_update" == "true" ]; then
  intro "Performing forced update - this process resets the machine-id and the update service. *A REBOOT WILL BE REQUIRED*"
  read -p "Do you need to perform a machine-id reset? Y/n (choosing \"Y\" can force a reboot): " response
    case $response in
      [Yy])
        reset_machine_id
        ;;
      *)
      ;;
    esac
  update_old_os
  if [[ $(grep "UPDATE_STATUS_REPORTING_ERROR_EVENT" $LOGFILE | wc -l) -gt 0 ]]; then
    echo "Found UPDATE_STATUS_REPORTING_ERROR_EVENT during update; shunting update-engine logs to stuntlog" | tee -a "$LOGFILE"
    intro "journalctl update-engine for last 2 hours"
    sudo journalctl --no-pager -u update-engine -S "2 hours ago" >> "$LOGFILE"
  fi
  if [[ $(grep "NO_UPDATE_AVAILABLE" $LOGFILE | wc -l) -gt 0 ]]; then
    echo "Found NO_UPDATE_AVAILABLE during update; Resetting machine-id again and doing double-update" | tee -a "$LOGFILE"
    reset_machine_id
    update_old_os
  fi
  outro
  endscript
  echo "EXITING"
  exit 0
fi
# End forced update procedure

# Alternating curl test
if [ "$curl_test" == "true" ]; then
  intro "Curl test starting"
  echo "Performing alternating curl test against S3 and SQS. This will run for $total_seconds_to_test seconds,"
  echo "and then quit automatically. Use ctrl+c to stop early."
  runme=true;
  run_sqs=true;
  now=$(date +%s) #time in seconds since epoch
  future=$(($now + $total_seconds_to_test))
  while $runme; do
    if [ "$run_sqs" == true ]; then
      echo $(date -u +"%b_%d_%y-%H:%M:%S") >> "$LOGFILE"
      echo "Testing connection to SQS: " | tee -a "$LOGFILE"
      expect "a 404 error."
      curl -Ssv -i -L --connect-timeout $seconds_between_tests "https://sqs.$AWS_REGION.amazonaws.com" >> "$LOGFILE"
      echo | tee -a "$LOGFILE"
      run_sqs=false;
      sleep 4;
    else
      echo $(date -u +"%b_%d_%y-%H:%M:%S") >> "$LOGFILE"
      echo "Testing connection to S3: " | tee -a "$LOGFILE"
      expect "a 403 error."
      curl -Ssv -i -L --connect-timeout $seconds_between_tests "https://sppcbu-va-images.s3.amazonaws.com" >> "$LOGFILE"
      echo | tee -a "$LOGFILE"
      run_sqs=true;
      sleep 4;
    fi
    if [[ $(date +%s) -ge $future ]]; then
      echo "$total_seconds_to_test seconds have elapsed. Quitting..."
      outro
      runme=false;
    fi
  done
  endscript
  exit 0
fi

### EXECUTE TESTS ###

intro "Retrieving list of files in home directory with ls -alh"
ls -alh /home/sailpoint/ >> "$LOGFILE"
outro

intro "Retrieving current working directory path with pwd"
expect "this to be /home/sailpoint/ but not a requirement"
pwd >> "$LOGFILE"
outro

key_passphrase_length=$(get_keyPassphrase_length)

intro "Retrieving config.yaml contents"
if [[ $key_passphrase_length -lt 1 ]]; then
  cat /home/sailpoint/config.yaml | sed -E "s/keyPassphrase: .*/keyPassphrase: <REMAINS UNENCRYPTED>/g" | sed "s/apiKey: .*/apiKey: <redacted>/g" >> "$LOGFILE"
else
  cat /home/sailpoint/config.yaml | sed -E "s/keyPassphrase: '[^:]*:[^:]*[^']*'.*/keyPassphrase: <redacted>/g" | sed "s/apiKey: .*/apiKey: <redacted>/g" >> "$LOGFILE"
fi
outro

perform_test "Is keyPassphrase length more than 0 characters?" "get_keyPassphrase_length" -gt 0 -lt 1 "configuration"
if [[ $key_passphrase_length -lt 1 ]]; then
  echo -e "     ${YELLOW}ACTION$RESETCOLOR: Check validity of config.yaml; keyPassphrase may still be unencrypted"
fi
outro

intro "Retrieving history of commands run on this session"
history >> "$LOGFILE"
outro

intro "Retrieving list of sudo commands run from journalctl"
sudo journalctl --no-pager _COMM=sudo  | grep -v -e "pam_unix(sudo:session)" >> "$LOGFILE"
outro

intro "This machine's IP address is: $IPADDR."
ip_result=$(is_ip_private $IPADDR)
if [[ $ip_result == 1 ]]; then
  echo -e "${YELLOW}WARNING:$RESETCOLOR This IP address appears to be public. It may have difficulty communicating with an internal DNS"
  echo -e "and may also be exposed to the public Internet - ${YELLOW}SailPoint recommends against exposing VAs to the edge of your network.$RESETCOLOR"
  add_test_result "networking" "Error handler" "warn" "$1"
  ((warnings++))
else
  perform_test "Is IP address ($IPADDR) private?" "is_ip_private $IPADDR" -eq 0 -eq 1 "networking"
fi
outro

perform_test "Does kernel version name report flatcar?" "uname -a | grep flatcar | wc -m" -gt 6 -eq 0 "system"
echo "uname output: $(uname -a)" >> "$LOGFILE"
outro

# CS0239311 - If left unsynced, could result in excessive message processing times
ntp_result=$(ntp_sync)
perform_test "Does timedatectl show NTP time is synced?" "ntp_sync" -eq 0 -ne 0 "configuration"
if [[ $ntp_result != 0 ]]; then
  echo -e "     ${YELLOW}ACTION REQUIRED:$RESETCOLOR Test for NTP sync failed. To configure NTP, see the following link: " | tee -a "$LOGFILE"
  echo -e "     https://documentation.sailpoint.com/saas/help/va/requirements_va.html#connecting-the-va-to-a-local-ntp-server" | tee -a "$LOGFILE"
fi
outro

intro "Retrieving OS reboot history"
last reboot >> "$LOGFILE"
outro

intro "Retrieving OS Uptime"
expect "this VA to have been restarted recently if it is having issues."
uptime >> "$LOGFILE"
outro

intro "Retrieving environment variables"
env >> "$LOGFILE"
outro

intro "Retrieving OpenJDK version from ccg"
expect "this version of java to be 11.0.14 or higher and not 1.8.x"
if test -f /home/sailpoint/log/worker.log; then
  echo "Grep in worker.log: $(grep -a openjdk /home/sailpoint/log/worker.log | tail -1)" >> "$LOGFILE"
elif test -f /home/sailpoint/log/ccg-start.log; then
  echo "Grep in ccg-start.log: $(grep -a "openjdk version" /home/sailpoint/log/ccg-start.log | tail -1)" >> "$LOGFILE"
elif test -f /home/sailpoint/log/ccg.log; then
  echo "Grep in ccg.log: $(grep -iE 'OpenJDK_64-Bit_Server_VM' /home/sailpoint/log/ccg.log | awk -F'OpenJDK_64-Bit_Server_VM' '{if (NF>1) {match($2, /[0-9]+\.[0-9]+\.[0-9]+/, version); if (version[0] != "") print version[0]}}' | tail -n1)" >> "$LOGFILE"
else
  echo -e "${YELLOW}WARNING:$RESETCOLOR Unable to find any log files to grep!"
fi
outro

intro "Checking for ip.list and retrieving contents of file"
if test -f /home/sailpoint/ip.list; then
  cat /home/sailpoint/ip.list >> "$LOGFILE"
elif [[ "$do_fixup" == true ]]; then
  echo -e "${INFO}INFO:$RESETCOLOR Resetting IP information in /home/sailpoint/ip.list file, then restarting services..."
  $(ip addr show label 'e[a-z][a-z][0-9]*' | grep -Po 'inet[6]* \K[\w.:]+' > /home/sailpoint/ip.list && sudo systemctl restart va_agent && sudo systemctl restart charon)
else
  echo -e "${YELLOW}WARNING:$RESETCOLOR ip.list file is missing. See KB article: https://sailpoint.service-now.com/kb?id=kb_article_view&sysparm_article=KB0019278" 
fi

intro "Checking for existence of files in /opt/sailpoint/share/"
perform_test "Does /opt/sailpoint/share/bin/ contain va-bootstrap?" "if [ -f \"/opt/sailpoint/share/bin/va-bootstrap\" ]; then echo 0; else echo 1; fi" -eq 0 -ge 1 "system"
perform_test "Does /opt/sailpoint/share/data/ contain a cluster_key.p12 file?" "if [ -f \"/opt/sailpoint/share/data/cluster_key.p12\" ]; then echo 0; else echo 1; fi" -eq 0 -ge 1 "system"
perform_test "Does /opt/sailpoint/share/ contain a ./jobs/ directory?" "if [ -d /opt/sailpoint/share/jobs/ ]; then echo 0; else echo 1; fi" -eq 0 -ge 1 "system"
outro

if test -f /etc/profile.env; then
  intro "Retrieving profile.env"
  expect "the file to exist. If proxy is a concern, have customer confirm settings."
  cat /etc/profile.env >> "$LOGFILE" 2>&1
  outro
fi

if test -f /etc/systemd/system.conf.d/10-default-env.conf; then
  intro "Retrieving 10-default-env.conf"
  expect "the file to exist. If proxy is a concern, have customer confirm settings."
  cat /etc/systemd/system.conf.d/10-default-env.conf >> "$LOGFILE" 2>&1
  outro
fi

intro "Retrieving docker.env"
expect "proxy references in docker.env. Remove references to proxy if proxying is a concern."
cat /home/sailpoint/docker.env >> "$LOGFILE"
outro

static_network_file_dhcp_yes=false #set default false as very few use dhcp=yes in static.network

if test -f /etc/systemd/network/static.network; then
  intro "Retrieving the static.network file"
  expect "individual DNS entries to be on separate lines beginning with 'DNS'."
  expect "the IP address to include CIDR notation."
  cat /etc/systemd/network/static.network >> "$LOGFILE"
if grep -q "DHCP=yes" /etc/systemd/network/static.network; then #check if static.network actually requests DHCP
    static_network_file_dhcp_yes=true
  else
    static_network_file_dhcp_yes=false
  fi
  outro
fi

intro "Retrieving the resolv.conf file"
expect "DNS entries to match those in static.network, if it exists."
cat /etc/resolv.conf >> "$LOGFILE"
outro

# tests for proxy.yaml
if test -f $PROXY_FILE_PATH ; then
  intro "Retrieving the proxy config"
  cat $PROXY_FILE_PATH >> "$LOGFILE"

  # CS0363011
  if [[ $(grep "no_proxy" $PROXY_FILE_PATH | wc -l) -ge 1 ]]; then
    perform_test "Is the value held in 'no_proxy' surrounded by double quotes?" "no_proxy_double_quotes" -eq 0 -ge 1 "configuration"
    perform_test "Is the value held in 'no_proxy' properly formatted and contains no commas?" "check_no_proxy_validate_format" -eq 0 -ge 1 "configuration"
  fi
  outro
fi

intro "Retrieving OS information from this system and Flatcar site"
expect "Version is the same as stable channel's most recent: https://www.flatcar.org/releases#stable-release"
scraped_flatcar_version=$(get_flatcar_current_version)
perform_test "Is current OS version the same as the one pulled from the Flatcar site?" "get_flatcar_current_version" "==" "$RUNNING_FLATCAR_VERSION" "!=" "$RUNNING_FLATCAR_VERSION" "system"
echo "OS version on this system: $RUNNING_FLATCAR_VERSION" >> "$LOGFILE"
echo "Stable OS version on Flatcar site: $scraped_flatcar_version" >> "$LOGFILE"
outro

intro "Checking if there's a Flatcar OS update waiting to be installed"
update_engine_status=$(get_update_engine_status 2>&1)
if [ $? -ne 0 ]; then             # the command get_update_engine_status_call timed out, so restart the service and try again
  sudo systemctl restart update-engine
  status_text=$(get_update_engine_status 2>&1)
fi
if [[ $(echo $update_engine_status | grep "UPDATE_STATUS_UPDATED_NEED_REBOOT" ) ]]; then
  echo -e "${INFO}INFO$RESETCOLOR: An OS update is waiting; please reboot." | tee -a $LOGFILE
  ADD_REBOOT_MESSAGE=true
else
  echo -e "${INFO}INFO$RESETCOLOR: Current update-engine status: $update_engine_status" | tee -a $LOGFILE
fi
outro

intro "Testing OS version"
perform_test "The OS version must not be 2345.x.y." "detect_old_os_version" -eq 0 -eq 1 "system"
outro


intro "Retrieving CPU information"
expect "the number of CPU(s) to be >= 4 CPUs. This is from AWS m4.large specs."
perform_test "Is number of CPUs greater than or equal to 4?" "get_lscpu_num_cpus" ">" 3 "<" 4 "system"
lscpu >> "$LOGFILE"
outro

intro "Retrieving total RAM"
expect "the RAM to be >= 15Gi (approx 16GB). This is from AWS m4.large specs."
perform_test "Is amount of RAM at least 16GB?" "get_free_total_mem_in_gb" -ge 15 -lt 15 "system"
free -h >> "$LOGFILE"
outro

intro "Retrieving current process list with 'top'"
get_top_output >> "$LOGFILE"
outro

intro "Network list for all adapters"
expect "one of two adapters to exist: ens160 or eth0. If canal is enabled, tun0 should be in this list as well."
if [[ "$IS_CANAL_ENABLED" == true ]]; then
  expect "that tun0 exists and is routable."
fi
networkctl list >> "$LOGFILE"
outro

intro "Network information for main adapter"
expect "information from resolv.conf/static.network/etc. to match up with what you find for the main adapter"
adapter_name=$(get_network_adapter_name)
echo "Network adapter name: $adapter_name" >> "$LOGFILE"
networkctl status $adapter_name >> "$LOGFILE"
outro

if [[ "$IS_CANAL_ENABLED" == true ]]; then
  expect "tun0 adapter to be in a 'routable (configuring)' state, and to show the online state as 'online'."
  networkctl status tun0 >> "$LOGFILE" 2>&1
fi

intro "Retrieving networking check in charon.log"
expect "all endpoints to have 'PASS' after their name"
perform_test "Post charon networking test, do all endpoints include PASS?" "get_charon_network_test_line | grep ERROR | wc -l" -eq 0 -ne 0 "networking"
get_charon_network_test_line | sed -n 's/.*Networking check results:\\n//;s/\\n/\n/gp' | awk -F'"' '{print $1}' >> "$LOGFILE"
outro

intro "Testing direct connection to regional Secure Tunnel servers"
expect "tests below to pass for every IP. On failure(s), ask if DPI (Deep Packet Inspection) or any variation is decrypting traffic from the VAs"
if [[ $PODNAME == *"useast1"* ||  $PODNAME == *"cook"* || $PODNAME == *"fiji"* || $PODNAME == *"uswest2"* || $PODNAME == *"cacentral1"* ]]; then
  # us-east-1 PODNAMEs contain: useast1 cook fiji uswest2 cacentral1
  echo "Using us-east-1 endpoints: " | tee -a "$LOGFILE"
  perform_test "Canal Server Connection Test to IP: 52.206.130.59" "canal_connection_test 52.206.130.59" -gt 4 -eq 0 "networking"
  outro
  perform_test "Canal Server Connection Test to IP: 52.206.133.183" "canal_connection_test 52.206.133.183" -gt 4 -eq 0 "networking"
  outro
  perform_test "Canal Server Connection Test to IP: 52.206.132.240" "canal_connection_test 52.206.132.240" -gt 4 -eq 0 "networking"
  outro
elif [[ $PODNAME == *"eucentral1"* ]]; then
  # eu-central-1 PODNAMEs contain: eucentral1
  echo "Using eu-central-1 endpoints: " | tee -a "$LOGFILE"
  perform_test "Canal Server Connection Test to IP: 35.157.132.22" "canal_connection_test 35.157.132.22" -gt 4 -eq 0 "networking"
  outro
  perform_test "Canal Server Connection Test to IP: 35.157.185.79" "canal_connection_test 35.157.185.79" -gt 4 -eq 0 "networking"
  outro
  perform_test "Canal Server Connection Test to IP: 35.157.251.228" "canal_connection_test 35.157.251.228" -gt 4 -eq 0 "networking"
  outro
elif [[ $PODNAME == *"euwest2"* ]]; then
  #eu-west-2 PODNAMEs contain: euwest2
  echo "Using eu-west-2 endpoints: " | tee -a "$LOGFILE"
  perform_test "Canal Server Connection Test to IP: 18.130.210.174" "canal_connection_test 18.130.210.174" -gt 4 -eq 0 "networking"
  outro
  perform_test "Canal Server Connection Test to IP: 18.130.148.201" "canal_connection_test 18.130.148.201" -gt 4 -eq 0 "networking"
  outro
  perform_test "Canal Server Connection Test to IP: 35.178.220.78" "canal_connection_test 35.178.220.78" -gt 4 -eq 0 "networking"
  outro
elif [[ $PODNAME == *"apsoutheast2"* ]]; then
  #apac PODNAMEs contain: apsoutheast2
  echo "Using ap-southeast-2 endpoints: "| tee -a "$LOGFILE"
  perform_test "Canal Server Connection Test to IP: 52.65.42.92" "canal_connection_test 52.65.42.92" -gt 4 -eq 0 "networking"
  outro
  perform_test "Canal Server Connection Test to IP: 13.55.78.212" "canal_connection_test 13.55.78.212" -gt 4 -eq 0 "networking"
  outro
  perform_test "Canal Server Connection Test to IP: 3.24.127.50" "canal_connection_test 3.24.127.50" -gt 4 -eq 0 "networking"
  outro
elif [[ $IS_ORG_FEDRAMP == true ]]; then
  #FEDRAMP
  echo "FedRAMP org detected - Canal servers not supported"| tee -a "$LOGFILE"
  outro
else
  echo "Unable to find appropriate canal server test with PODNAME: $PODNAME" >> "$LOGFILE"
  outro
fi

intro "Retrieving contents of /home/sailpoint/hosts.yaml"
if [[ -e "/home/sailpoint/hosts.yaml" ]]; then
  cat /home/sailpoint/hosts.yaml >> "$LOGFILE"
else
  echo "INFO - /home/sailpoint/hosts.yaml not found" >> "$LOGFILE"
  echo -e "${INFO}INFO$RESETCOLOR: hosts.yaml not found"
fi
outro

# intro "Getting RAM stats from ccg container"
# expect "RAM to be at least 8GB for sandbox, and typically 16GB for prod applications"
# sudo docker stats ccg --no-stream | awk 'NR==2' | awk '{ print strftime("[%Y-%m-%d %H:%M:%S]"), $0 }' >> "$LOGFILE" ; 
# outro

intro "Retrieving contents of /etc/hosts from host"
expect "entries to match the /etc/hosts from ccg in the next section."
cat /etc/hosts >> "$LOGFILE"
outro



intro "Retrieving contents of /etc/hosts from ccg container."
IS_CCG_RUNNING=$(check_container_running "ccg")
if [[ $IS_CCG_RUNNING == true ]]; then
  sudo docker exec ccg cat /etc/hosts >> "$LOGFILE" 2>&1
else
  echo -e "${YELLOW}WARNING$RESETCOLOR: CCG container is not running."
  echo "WARNING: CCG container is not running." >> "$LOGFILE"
fi
outro


intro "Retrieving contents of /opt/sailpoint/ccg/lib/custom from ccg container"
expect "JDBC driver jar files. Make sure they are populating here if required."
sudo docker exec ccg ls -al /opt/sailpoint/ccg/lib/custom >> "$LOGFILE" 2>&1
outro

if [[ "$do_fixup" == true ]]; then
  intro "This step disables esx_dhcp_bump if needed"
  expect "any output stating this was removed/disabled. If there is, be sure to do a sudo reboot."
  if [[ "$static_network_file_dhcp_yes" == false ]]; then
    sudo systemctl disable esx_dhcp_bump >> "$LOGFILE" 2>&1
  else
    echo "Not disabling esx_dhcp_bump because static.network has DHCP=yes. Disable manually with 'sudo systemctl disable esx_dhcp_bump' command if DHCP is not in use, then reboot the VA."  >> "$LOGFILE" 2>&1
  fi
  outro
fi

# CS0363017
intro "Checking for the existence of override.conf"
expect "this file not to exist"
if [[ -e /etc/systemd/system/update-engine.service.d/override.conf ]]; then
  if [[ "$do_fixup" == true ]]; then
    echo -e "${INFO}INFO:$RESETCOLOR override.conf found and fixup enabled. Attempting removal..." | tee -a "$LOGFILE"
    sudo rm /etc/systemd/system/update-engine.service.d/override.conf | tee -a "$LOGFILE"
  fi
  echo -e "${YELLOW}ACTION:$RESETCOLOR override.conf file found, but fixup option is not enabled. Rerun script with fixup (-f) to remove this file." | tee -a "$LOGFILE"
else
  echo "File not found, as expected" | tee -a "$LOGFILE"
fi
outro

if [[ "$get_ssl_conf" == true ]]; then
  intro "Retrieving openssl config file contents"
  cat /etc/ssl/openssl.cnf >> "$LOGFILE"
  outro
fi

intro "Retrieving list of all SSL certs in $CERT_DIRECTORY"
ls -alh $CERT_DIRECTORY >> "$LOGFILE" 2>&1
outro

if [[ $( ls -1 /home/sailpoint/certificates | wc -l ) -gt 0 ]]; then
  intro "Test all custom certificates in $CERT_DIRECTORY against openssl"
  cert_tester
  outro
fi

if [[ "$do_fixup" == true ]]; then
  intro "This step updates all SSL certificates in /etc/ssl/certs"
  expect "this section to be blank; we only catch errors here."
  sudo /usr/sbin/update-ca-certificates > /dev/null 2>> "$LOGFILE" # stdout to /dev/null, catch only errors
  outro
fi

intro "Retrieving list of all SSL certs in /etc/ssl/certs"
expect "updated date for most certs to be very recent if update-ca-certificates has been used."
ls -alh /etc/ssl/certs >> "$LOGFILE" 2>&1
outro

# Tenant URL: "https://$ORGNAME.$ISC_DOMAIN", 
# SQS example: "https://sqs.$AWS_REGION.amazonaws.com", 
# API: "https://$ORGNAME.$ISC_ACCESS"

# TODO: use ping to check if sites are resolving first, and if successful, then execute curl. The --connect-timeout option isn't working as anticipated.
if [[ $IS_IAI_VA == true ]]; then
  intro "External connectivity: Connection test to launchdarkly (https://app.launchdarkly.com); ignores chain, outputs SSL info and HTTP status"
  curl -vvvIik --connect-timeout $seconds_between_tests https://app.launchdarkly.com >> "$LOGFILE" 2>&1
  perform_test "Curl test to launchdarkly; expect a result of 405" "curl -vvvIik \"https://app.launchdarkly.com\" 2>&1 | grep \"405\" | wc -l" -gt 0 -eq 0 "networking"
  outro
fi

#v2.3.8 - FedRAMP supports updated VA pairing
intro "External connectivity: Connection test to the va-activation endpoint to get a code (blank response is ok)"
curl -Svv -k "https://va-activation-global.secure-api.infra.identitynow.com/activation/code" >> "$LOGFILE" 2>&1 || true
outro

intro "External connectivity: Connection test for SQS (https://sqs.$AWS_REGION.amazonaws.com)"
curl -Ssv -i -L -vv --connect-timeout $seconds_between_tests "https://sqs.$AWS_REGION.amazonaws.com" >> "$LOGFILE" 2>&1
outro
perform_test "Curl test to SQS; expect a result of 404" "curl -i --connect-timeout $seconds_between_tests \"https://sqs.$AWS_REGION.amazonaws.com\" 2>&1 | grep \"404 Not Found\" | wc -l" -gt 0 -eq 0 "networking"
outro

if [[ $ORGNAME != "mytestorg" ]]; then # Skip test when temporary config.yaml in place
  intro "External connectivity: Connection test for main URL (expected failure on vanity) https://$ORGNAME.$ISC_DOMAIN"
  curl -Ssv -i --connect-timeout $seconds_between_tests "https://$ORGNAME.$ISC_DOMAIN" >> "$LOGFILE" 2>&1
  outro
  perform_test "Curl test to IdentityNow org; expect a result of 302" "curl -i --connect-timeout $seconds_between_tests \"https://$ORGNAME.$ISC_DOMAIN\" 2>&1 | grep -e 'HTTP/2 302\|HTTP/1.1 302 Found' | wc -l" -gt 0 -eq 0 "networking" 
  outro

  if [[ $IS_ORG_FEDRAMP == true ]]; then
    intro "External connectivity: Connection test for https://$ORGNAME.$ISC_ACCESS"
    curl -Ssv -i -L --connect-timeout $seconds_between_tests "https://$ORGNAME.$ISC_ACCESS" >> "$LOGFILE" 2>&1
    outro
    perform_test "Curl test to the tenant API; expect a result of 404" "curl -i --connect-timeout $seconds_between_tests \"https://$ORGNAME.$ISC_ACCESS\" 2>&1 | grep \"404\" | wc -l" -gt 0 -eq 0 "networking"
    outro
  fi
fi

intro "External connectivity: Connection test for DynamoDB (https://dynamodb.$AWS_REGION.amazonaws.com)"
curl -Ssv -i -L --connect-timeout $seconds_between_tests "https://dynamodb.$AWS_REGION.amazonaws.com" >> "$LOGFILE" 2>&1
outro
perform_test "Curl test to DynamoDB; expect a result of 200" "curl -i --connect-timeout $seconds_between_tests \"https://dynamodb.$AWS_REGION.amazonaws.com\" 2>&1 | grep \"HTTP/1.1 200 OK\" | wc -l" -gt 0 -eq 0 "networking"
outro

if [[ $IS_ORG_FEDRAMP == false ]]; then
  intro "External connectivity: Connection test for starport bucket in S3"
  curl -vvv "https://prod-us-east-1-starport-layer-bucket.s3.$AWS_REGION.amazonaws.com" >> "$LOGFILE" 2>&1
  perform_test "Curl test to starport s3 bucket" "curl -vvv --connect-timeout $seconds_between_tests \"https://prod-us-east-1-starport-layer-bucket.s3.$AWS_REGION.amazonaws.com\" 2>&1 | grep \"403 Forbidden\" | wc -l" -gt 0 -eq 0 "networking"
else
  intro "External connectivity: Connection test for starport bucket in FedRAMP S3"
  curl -vvv "https://s3-fips.$AWS_REGION.amazonaws.com" >> "$LOGFILE" 2>&1
  perform_test "Curl test to FedRAMP s3" "curl -vvv --connect-timeout $seconds_between_tests \"https://s3-fips.$AWS_REGION.amazonaws.com\" 2>&1 | grep \"307 Temporary\" | wc -l" -gt 0 -eq 0 "networking"
fi
outro

intro "Checking active network ports using netstat."
sudo netstat -pan -A inet,inet6 | grep -v ESTABLISHED >> "$LOGFILE" 2>&1
outro

intro "Using the ss utility to list open ports"
ss -plno -A tcp,udp,sctp >> "$LOGFILE"
outro

intro "Display network (tcp) statistics"
expect "the number of failed connection attempts to be less than 100. If more, consider a packet capture."
echo "failed connection attempts:   high numbers indicate issues establishing connections, could be network, resource limits or firewall rules" >> "$LOGFILE"
echo "connection resets received:   high numbers indicate problems with remote servers or network paths" >> "$LOGFILE" 
echo "segments retransmitted:       indicates possible packet loss" >> "$LOGFILE"
echo "bad segments received:        a sign of network issues" >> "$LOGFILE"
echo "resets sent:                  high numbers indicate problems with the system rejecting connections" >> "$LOGFILE"
echo $DIVIDER | tee -a "$LOGFILE"
sudo netstat -st >> "$LOGFILE" 2>&1
outro

intro "Performing OpenSSH version test"
perform_test "Output from 'ssh -V' is $openssh_version; expect it to be 10.0 or higher" "test_openssh_version" -eq 0 -ne 0 "networking"
outro 

intro "Retrieving information on how system is using DNS for each link"
systemd-resolve --status >> "$LOGFILE"
outro

if [ "$do_ping" = true ]; then
  intro "Pinging IdentityNow tenant"
  ping -c 5 -W 2 $ORGNAME.identitynow.com >> "$LOGFILE"
  outro
fi

if [ "$do_traceroute" = true ]; then
  intro "Collecting traceroute to SQS... (this may take a moment; please be patient)"
  traceroute sqs.$AWS_REGION.amazonaws.com >> "$LOGFILE"
  outro
fi

intro "Retrieving additional routing information from ip route show"
ip route show >> "$LOGFILE"
outro

# Only gather log snippets if we're not getting all logs due to -n flag
if [[ "$gather_logs" != true ]]; then
  intro "Retrieving ccg.log errors - latest 30 errors"
  expect "recent datestamps. Some logs might be old and no longer pertinent. Expect no keystore.jks or 'decrypter' errors. These signify a keyPassphrase issue."
  cat /home/sailpoint/log/ccg.log | grep stacktrace | tail -n30 >> "$LOGFILE" 2>&1
fi

intro "Checking Charon version"
expect "Charon version should be higher than $CHARON_MINIMUM_VERSION"
current_charon=$(get_current_image_tag charon) 
perform_test "Is charon version higher than $CHARON_MINIMUM_VERSION?" "if [[ $current_charon > $CHARON_MINIMUM_VERSION ]]; then echo true; fi" "==" "true" "==" "false" "system"
echo "Current charon version is $current_charon" >> "$LOGFILE" 2>&1

if [ -n "$current_charon" ] && [ "$current_charon" -lt "$CHARON_MINIMUM_VERSION" ]; then
  echo "Current version of charon is too old." >> "$LOGFILE" 2>&1
  if [ "$do_fixup" == true ]; then
    echo "Restarting container to help update" | tee -a "$LOGFILE"
    sudo systemctl restart va_agent 2>&1 | tee -a "$LOGFILE"
    echo "VA Agent restart, waiting 30 seconds before restarting Charon" | tee -a "$LOGFILE"
    sleep 30
    sudo systemctl restart charon 2>&1 | tee -a "$LOGFILE"
    echo "Charon restarted. Please monitor its logs for connection or authentication errors" | tee -a "$LOGFILE"
    endscript
    exit 0
  else
    echo "Charon container is running a older build that may fail to update."  >> "$LOGFILE" 2>&1
    echo "Rerun this script with -f to automatically restart it."  >> "$LOGFILE" 2>&1
    echo "Otherwise, run 'sudo systemctl restart charon' or reboot the entire VA".  >> "$LOGFILE" 2>&1
  fi
fi
outro

expect "the CCG image to be updated: it should be less than 1 month old."
docker_images=$(echo "$DOCKER_IMAGES_OUTPUT" | sort)
echo -e "$docker_images" >> "$LOGFILE"
if echo -e "$docker_images" | grep -q "sailpoint/charon"; then
  : # charon is present
else
  echo "ERROR: Charon image is missing." | tee -a "$LOGFILE"
  if [ "$do_fixup" == true ]; then
    echo "Attempting to fix missing images" | tee -a "$LOGFILE"
    fix_missing_images >> "$LOGFILE" 2>&1
  else
    echo "Charon image is missing and fixup is disabled. Rerun script with fixup (-f) to attempt repair." | tee -a "$LOGFILE"
  fi
fi
outro

expect "the following four (4) processes to be running: ccg, va_agent, charon, and va."
perform_test "Is ccg running?" "check_container_running \"ccg\"" "==" "true" "==" "false" "system"
outro
perform_test "Is va_agent running?" "check_container_running \"va_agent\"" "==" "true" "==" "false" "system"
outro
perform_test "Is charon running?" "check_container_running \"charon\"" "==" "true" "==" "false" "system"
outro
perform_test "Are either fluent or vector running?" "check_container_running \"fluent\|vector\"" "==" "true" "==" "false" "system"
outro
if [[ "$IS_CANAL_ENABLED" == true ]]; then
  expect "an additional service to be running when Secure Tunnel is enabled: canal"
  perform_test "Is canal running?" "check_container_running \"canal\"" "==" "true" "==" "false" "system"
  outro
fi

intro "Retrieving ccg container configuration from /proc/meminfo"
sudo docker exec ccg ls /proc/meminfo | xargs cat >> "$LOGFILE"
outro

intro "Retrieving list of running systemd services"
systemctl list-units --type=service --state=running >> "$LOGFILE"
outro

intro "Retrieving list of exited systemd services"
systemctl list-units --type=service --state=exited >> "$LOGFILE"
outro

intro "Retrieving systemd service configuration file: charon"
expect "the file to exist, and contains a valid docker ECR address compared to the docker images list above."
cat /etc/systemd/system/charon.service >> "$LOGFILE"
outro

intro "Retrieving systemd service configuration file: ccg"
expect "the file to exist, and contains a valid docker ECR address compared to the docker images list above."
cat /etc/systemd/system/ccg.service >> "$LOGFILE"
outro

intro "Retrieving systemd service configuration file: va_agent"
expect "the file to exist, and contains a valid docker ECR address compared to the docker images list above."
cat /etc/systemd/system/va_agent.service >> "$LOGFILE"
outro

intro "Retrieving systemd service configuration file: fluent"
expect "the file to exist, and contains a valid docker ECR address compared to the docker images list above."
cat /etc/systemd/system/fluent.service >> "$LOGFILE"
outro

intro "Retrieving systemd service configuration file: relay"
expect "the file to exist, and contains a valid docker ECR address compared to the docker images list above."
cat /etc/systemd/system/relay.service >> "$LOGFILE"
outro

intro "Retrieving systemd service configuration file: toolbox"
expect "the file to exist, and contains a valid docker ECR address compared to the docker images list above."
cat /etc/systemd/system/toolbox.service >> "$LOGFILE"
outro

intro "Retrieving the service-config dependencies list"
expect "the subset to contain up-to-date version information for containers required by services."
cat /opt/sailpoint/share/service-config.json | jq .dependencies >> "$LOGFILE"
outro

if [ "$IS_ORG_FEDRAMP" = true ]; then
  intro "'fips' string should exist in grub.cfg"
  perform_test "Does the 'fips' string exist in the grub.cfg file?" "grep fips /oem/grub.cfg | wc -l" -eq 1 -lt 1 "system"
  outro
fi

if [[ "$virt_host" == "microsoft" ]]; then
  intro "Azure-hosted VAs must have the waagent service masked, not just disabled."
  perform_test "Is waagent masked?" "sudo systemctl status waagent | grep "masked" | wc -l" -eq 1 -lt 1 "system"
  outro
fi

if [[ "$IS_CANAL_ENABLED" == true ]]; then
  intro "Retrieving systemd service configuration file: canal"
  expect "the file to exist, and contains a valid docker ECR address compared to the docker images list above."
  cat /etc/systemd/system/canal.service >> "$LOGFILE"
  outro
fi

if test -f /etc/systemd/system/esx_dhcp_bump.service; then
  intro "Retrieving systemd service configuration file: esx_dhcp_bump"
  cat /etc/systemd/system/esx_dhcp_bump.service >> "$LOGFILE"
  outro
fi

intro "Retrieving partition table info"
expect "total disk space under \"SIZE\". Should be ~128GB or more."
expect "one sda<#> to be TYPE 'part' and RO '0'. This means the PARTition is writable."
lsblk -o NAME,SIZE,FSTYPE,FSSIZE,FSAVAIL,FSUSE%,MOUNTPOINT,TYPE,RO >> "$LOGFILE"
outro

intro "Retrieving disk usage stats"
expect "the root filesystem to be less than 15% full (typically sda9 or similar). Potential a debug setting was enabled long-term."
expect "the filesystem types to be devtmpfs, tmpfs, vfat, overlay, or ext4, and NOT btrfs"
df -Th >> "$LOGFILE"
outro

intro "Checking if Root filesystem has at least $ROOT_FS_MINIMUM_FREE_KB kilobytes free"
expect "Root filesystem should have at $ROOT_FS_MINIMUM_FREE_KB free"
root_free_kb=$(df -k | grep " /$" | awk '{print $4}')
echo "Root FS has $root_free_kb KB free" >> "$LOGFILE" 2>&1
if [ "$root_free_kb" -lt "$ROOT_FS_MINIMUM_FREE_KB" ]; then
  echo "Root FS has only $root_free_kb"  >> "$LOGFILE"
  if [ "$do_fixup" == true ]; then
    clean_non_current_images >> "$LOGFILE" 2>&1
  else
    echo "VA is low on disk space. Re-run this script with -f to clean unused container images to free space" >> "$LOGFILE" 2>&1
    echo "You can also truncate the ccg logfile, but only do so if you do not need the data in it." >> "$LOGFILE" 2>&1
    echo "To truncate the ccg log, run 'truncate -s 0 /home/sailpoint/log/ccg.log" >> "$LOGFILE" 2>&1
  fi
fi
outro

intro "Retrieving disk usage paths"
expect "most files to be less than 1GB. Log files can be significantly larger, but shouldn't exceed 1GB each."
du -h /home/sailpoint/ >> "$LOGFILE"
outro

intro "Retrieving list of large files"
expect "most files to be less than 1MB. Log files can be significantly larger, but shouldn't exceed 1GB each."
find ~/log/ -xdev -type f -size +100M | xargs ls -lh | sort -k5,5 -h -r >> "$LOGFILE"
outro

perform_test "Are more than 100 inodes available on the main partition?" "(df -i | awk -v partition=\"$MAIN_PARTITION\" '\$1 == partition {print \$4}' | tail -n1)" -gt 100 -lt 100 "system" 
outro 

intro "Retrieving number and list of pending jobs."
num_pending_jobs=$(ls /opt/sailpoint/workflow/jobs/ | wc -l)
echo "$num_pending_jobs pending jobs in the directory." >> "$LOGFILE"
ls -al /opt/sailpoint/workflow/jobs >> "$LOGFILE"
outro

expect "this to have fewer than 20 completed jobs. If lots of jobs are > 1 week old, run: sudo rm -rf /opt/sailpoint/share/jobs/* && sudo reboot"
perform_test "Does /opt/sailpoint/share/jobs have fewer than 20 jobs?" "get_num_share_jobs" -lt 20 -gt 19 "system"
outro

expect "this to have fewer than 20 workflow jobs. If lots of jobs are > 1 week old, run: sudo rm -rf /opt/sailpoint/workflow/jobs/* && sudo reboot"
perform_test "Does /opt/sailpoint/workflow/jobs have fewer than 20 jobs?" "get_num_workflow_jobs" -lt 20 -gt 19 "system"
outro

if [[ "$IS_CANAL_ENABLED" == true ]]; then
  echo "$DIVIDER"
  intro "The following tests and data gathering are only run if Secure Tunnel config has been enabled"
  echo
  intro "Retrieving the canal config file @/opt/sailpoint/share/canal/client.conf"
  cat /opt/sailpoint/share/canal/client.conf >> "$LOGFILE"
  outro

  perform_test "The canal.env file should exist" "canalenv_exists" -eq 0 -eq 1 "system"
  outro

  perform_test "The canal-hc.log file should exist" "canal-hc_log_exists" -eq 0 -eq 1 "system"
  does_canal_hc_log_exist=$(canal-hc_log_exists)
  if [[ "$does_canal_hc_log_exist" == 1 ]]; then
    if [[ "$do_fixup" == true ]]; then
      echo "Creating canal-hc.log" | tee -a "$LOGFILE"
      touch /home/sailpoint/log/canal-hc.log | tee -a "$LOGFILE"
    else
      echo -e "${YELLOW}ACTION:$RESETCOLOR File does not exist, but the option for automatic fixup" | tee -a "$LOGFILE"
      echo -e "is not enabled. Rerun STUNT with -f to create the canal-hc.log file" | tee -a "$LOGFILE"
    fi
  fi
  outro

  perform_test "The canal-hc.log file should contain 0 instances of the error message 'No such file or directory'" "canal_log_contains_FNF_string" -eq 0 -gt 0 "system"
  outro

  intro "Checking charon.log for successful canal setup"
  perform_test "Check charon.log for canal setup success message" 'grep -e "SUCCESS" -e "canal" /home/sailpoint/log/charon.log | tail -n1' "==" "Job SERVICE_SETUP fluent/ccg/relay/canal has FINISHED - result: SUCCESS" "==" "" "configuration" 

  grep -e "SUCCESS" -e "canal" /home/sailpoint/log/charon.log | tail -n1 >> "$LOGFILE"
  outro

  intro "Retrieving last 50 lines of canal service journal logs"
  sudo journalctl --no-pager -n50 -u canal >> "$LOGFILE"
  outro

  echo "*** Completed gathering extra data from Canal config."
  echo "$DIVIDER"
  echo
fi

intro "Gathering logrotate service info (/usr/lib/systemd/system/logrotate.service)"
cat /usr/lib/systemd/system/logrotate.service >> "$LOGFILE"
outro

intro "Gathering logrotate configuration info (/usr/share/logrotate/logrotate.conf)"
cat /usr/share/logrotate/logrotate.conf >> "$LOGFILE"
outro

intro "Checking for modified ccg java heap settings"
expect "file not found at location: $JAVA_OVERWRITES_FILE_PATH."
if [[ -e  $JAVA_OVERWRITES_FILE_PATH ]]; then
  echo -e "${INFO}INFO:$RESETCOLOR Found java heap settings have been manually set. Please follow compatibility guidelines." | tee -a "$LOGFILE"
  echo -e "https://community.sailpoint.com/t5/IdentityNow-Draft-Documents/Increasing-memory-usage-on-the-VA-Java-heap/ta-p/78766"
  cat $JAVA_OVERWRITES_FILE_PATH >> "$LOGFILE"
else
  echo -e "Overwrites file not found."
fi
outro

intro "Testing for decrypter errors in ccg.log"
perform_test "Check ccg.log for indicators of mismatched keyPassphrase" 'tail --bytes 10M /home/sailpoint/log/ccg.log | grep -i "An error occurred while decrypting the message" | wc -l' -eq 0 -gt 0 "configuration"
outro

intro "Retrieving last 25 lines of error logs from dmesg"
expect "this to be blank. Any kernel ring buffer or hv_netvsc (Hyper-V specific) messages likely reveal hardware-related errors."
dmesg | grep -i "error" | tail -n25 >> "$LOGFILE"
outro

intro "Retrieving last week of logrotate.service journal logs"
sudo journalctl --no-pager -u logrotate.service -S "1 week ago" | tee -a "$LOGFILE" | grep -q "error: unable to open /home/sailpoint/log/ccg.log.1 (read-only)" | wc -l
if [ $? -gt 0 ] && [ -e /home/sailpoint/log/ccg.log.1 ]; then
  if [[ "$do_fixup" == true ]]; then
    echo "Found corrupted logrotate cache. Attempting fixup." >> "$LOGFILE"
    echo "-----" >> /home/sailpoint/log/ccg.log.1 && sudo systemctl start logrotate.service
  else 
    echo -e "${YELLOW}ACTION: $RESETCOLOR 'No such file or directory' error found in logrotate.service journal logs," | tee -a "$LOGFILE"
    echo -e "but the option for automatic fixup is not enabled. Rerun STUNT with -f to attempt automatic correction." | tee -a "$LOGFILE"
  fi
fi
outro

intro "Retrieving last 2 days of logrotate.timer logs"
sudo journalctl --no-pager -u logrotate.timer -S "2 days ago" >> "$LOGFILE"
outro

intro "Retrieving last 50 lines of ccg journal logs"
sudo journalctl --no-pager -n50 -u ccg >> "$LOGFILE"
outro

intro "Retrieving last 50 lines of charon journal logs"
sudo journalctl --no-pager -n50 -u charon >> "$LOGFILE" 
outro

intro "Retrieving last 50 lines of va_agent journal logs"
sudo journalctl --no-pager -n50 -u va_agent >> "$LOGFILE"
outro

intro "Retrieving last 50 lines of otel_agent journal logs"
sudo journalctl --no-pager -n50 -u otel_agent >> "$LOGFILE"
outro

# CS0360097
intro "Retrieving last 2 hours of update-service (update-engine) journal logs"
sudo journalctl --no-pager -u update-engine -S "2 hours ago" >> "$LOGFILE" && sync
if grep -q "Unknown Omaha response status: error-internal" "$LOGFILE"; then
  echo -e "${YELLOW}WARNING:$RESETCOLOR Found 'Unknown Omaha response status: error-internal'"
  echo "This indicates a machine-id conflict between this instance and the update server."
  if [[ "$do_fixup" == true ]]; then
    echo "Removing the machine-id automatically. A REBOOT IS REQUIRED."
    sudo rm -f /etc/machine-id && ADD_REBOOT_MESSAGE=true
  else
    echo -e "${YELLOW}ACTION:$RESETCOLOR The machine-id must be reset, but the option for automatic fixup "
    echo "is not enabled. Re-run stunt with the '-f' flag."
  fi
fi
outro

intro "Retrieving last 50 lines of kernel journal logs"
sudo journalctl --no-pager -n50 -k >> "$LOGFILE"
outro

intro "Retrieving last 50 lines of network journal logs"
sudo journalctl --no-pager -n50 -u systemd-networkd >> "$LOGFILE"
outro

intro "Retrieving last 50 lines of docker-related journal logs"
sudo journalctl --no-pager -n50 -u docker >> "$LOGFILE"
outro

intro "Retrieving all dockerd journal logs from the last week"
sudo journalctl --no-pager -S "1 week ago" | grep dockerd >> "$LOGFILE"
outro

intro "Retrieving the last 1 hour of journal logs"
sudo journalctl --no-pager -S "1 hour ago" >> "$LOGFILE"
outro

endscript

all_test_results=$(output_all_tests_by_category)

echo $DIVIDER | tee -a "$LOGFILE"
echo "Testing summary"
echo $DIVIDER
echo -e "Tests passed:                           $GREEN $passes $RESETCOLOR"
echo -e "Tests failed:                           $RED $failures $RESETCOLOR"
echo -e "Test warnings:                          $YELLOW $warnings $RESETCOLOR"
echo
echo $DIVIDER

summary="*** Post test summary ***
Tests passed:   $passes 
Tests failed:    $failures 
Test warnings:  $warnings 

Tests:
$all_test_results"

# CS0360919
echo "$summary" > /tmp/summary_temp.txt
awk -v var="$(cat /tmp/summary_temp.txt)" '{gsub(/<SUMMARY_BLOCK>/, var)}1' $LOGFILE > temp && mv temp $LOGFILE && sync && rm /tmp/summary_temp.txt

if [ "$capture_journal" != true ] && [ "$gather_logs" == true ]; then
  # archive stuntlog, all logs in /home/sailpoint/log/ and config files in /home/sailpoint/ccg/
  intro "Gathering log files and ccg directory and zipping. 3"
  echo
  echo "*** NOTE: This file might be large depending on the lifespan and uptime of your VA. ***"
  echo
  zip -r $ZIPFILE $LOGFILE $LISTOFLOGS $CCGDIR
fi
if [ "$gather_logs" != true ] && [ "$capture_journal" == true ]; then
  # archive stuntlog and journal log
  echo "*** Gathering last 24hrs of systemd journal and archiving ***"
  sudo journalctl --no-pager -S "1 day ago" > /home/sailpoint/log/journal-$(date +%Y-%m-%d_%H:%M:%S).log && sync
  zip -r $ZIPFILE $LOGFILE /home/sailpoint/log/journal-* &&
  echo "*** Removing temporary systemd journal log *1*"
  rm /home/sailpoint/log/journal-*.log
fi
if [ "$gather_logs" == true ] && [ "$capture_journal" == true ]; then
  # archive stuntlog, journald log, all logs in /home/sailpoint/log/ and config files in /home/sailpoint/ccg/
  echo "*** Gathering last 24hrs of systemd journal, all log files, ccg directory, and archiving ***"
  sudo journalctl --no-pager -S "1 day ago" > /home/sailpoint/log/journal-$(date +%Y-%m-%d_%H:%M:%S).log && sync
  zip -r $ZIPFILE $LOGFILE $LISTOFLOGS $CCGDIR &&
  echo "*** Removing temporary systemd journal log *2*"
  rm /home/sailpoint/log/journal-*.log
fi
outro

if [ "$gather_logs" == true ] ; then #|| ([ "$gather_logs" == true && "$capture_journal" == true ])
  echo "Zipped to $ZIPFILE" | tee -a "$LOGFILE"
  # Remove the stuntlog file now that it's in the zip
  rm $LOGFILE
  echo
  echo -e $REDBOLDUL"*** Retrieve this zipped file, without renaming, and upload to your case:"
  echo -e $GREEN"${ZIPFILE} "
  echo
  echo -e $YELLOW"EXIT the VA, and use 'scp' from a shell/terminal to retrieve the zip file from this server."
  echo "Use the line created for you below at your local machine's shell/terminal:"
  echo 
  echo -e $CYAN"scp sailpoint@$IPADDR:$ZIPFILE ./ $RESETCOLOR"
else
  echo
  echo "*** RETRIEVE THE FILE WITHOUT RENAMING:"
  echo "*** ${LOGFILE} "
  echo "*** AND UPLOAD TO YOUR CASE."
  echo 
  echo -e $REDBOLDUL$DIVIDER
  echo -e $GREEN"File created: $LOGFILE"
  echo "Retrieve it and attach it to your case"
  echo -e $REDBOLDUL$DIVIDER
  echo 
  echo -e $YELLOW"EXIT the VA and use 'scp' from a shell/terminal to retrieve the log file from this server."
  echo -e "Use the line created for you below at your local machine's shell/terminal:"
  echo
  echo -e $CYAN"scp sailpoint@$IPADDR:$LOGFILE ./ $RESETCOLOR"
fi

if [ "$ADD_REBOOT_MESSAGE" == true ]; then
  echo -e "$YELLOW$DIVIDER $REDBOLDUL"
  echo -e "REBOOT IS REQUIRED; PLEASE RUN 'sudo reboot' NOW $RESETCOLOR"
  echo -e "$YELLOW$DIVIDER $RESETCOLOR"
fi
