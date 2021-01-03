#!/usr/bin/env sh

################################################################
# Automatic installation script for the larashed go-agent
#
# Usage: `curl -sSL %%URL_TO_SCRIPT | ENVIRONMENT_VARIABLES... sudo sh`
#
# Steps after each script call:
# 0. Check calling user is root and sudo is installed
# 1. Parse arguments
# Install-Workflow:
# 1. Check if wget or curl are present
# 2. Create tmp-dir
# 3. Get release information from api.github.com
#    for the project specified in $GITHUB_URL
# 4. Download the binary named $LINUX_BINARY_NAME in the release
# 5. Download the first file in the release, matching the pattern
#    $LINUX_BINARY_NAME.XYZ where XYZ can be any hash-algorithm command like sha256sum or
#    md5sum. The script will automatically try to run the
#    hash-command found after the last dot in the filename!
#    If the checksum file would be named e.g. binary.md5sum,
#    the binary would be compared using the md5sum command.
# 6. Compare binary checksum with downloaded checksum, exit on mismatch
# 7. Copy binary to destination, set permissions
# 8. Generate config files
# 9. Generate entry in /etc/sudoers
# 10. Generate systemd unit
################################################################

################################################################
# declaration of global variables
################################################################

# naming and download URLs
APP_NAME="Larashed monitoring agent"                                # Used in prints
APP_NAME_SLUG="larashed-agent"                                      # Used in service name etc. - DO NOT CHANGE
INSTALLER_VERSION="v1.0"                                            # Used in prints
GITHUB_URL="https://github.com/larashed/agent-go"                   # Used for download
LINUX_BINARY_NAME="agent_linux_amd64"                               # Name of the Linux binary inside each release
# installation Settings
BINARY_DESTINATION="/usr/local/bin/larashed-agent"                  # FHS conform path for of executable
CONFIG_FOLDER="/etc/larashed"                                       # FHS conform config path
CONFIG_FILE="larashed.conf"                                         # config filename inside $CONFIG_FOLDER
UNIX_USERNAME="larashed"                                            # Username for the agent-runuser
SYSTEMD_UNIT_PATH="/etc/systemd/system/$APP_NAME_SLUG.service"      # Path to the systemd unit file
SYSTEMD_RESTART_TIME="2"                                            # Time to wait after stop during systemd restart
SYSTEMD_TIMEOUT="10"                                                # Time to wait for agent to shutdown before force kill
# script parameters
SCRIPT_MODE="install"                                               # [install|uninstall] overridden with --uninstall
PERFORM_UPDATE=0                                                    # overridden with --update
VERBOSE="false"                                                     # overridden with -v
# variables for internal usage
CHOSEN_DOWNLOADER=""                                                # Will be either curl or wget
LATEST_LINUX_BINARY_URL=""                                          # Auto populated from GitHub API
LATEST_LINUX_BINARY_HASH_URL=""                                     # Auto populated from GitHub API
LATEST_LINUX_BINARY_VERSION=""                                      # Auto populated from GitHub API
TEMP_WORK_DIR=""                                                    # Populated by mktemp, needed for cleanup
RED='\033[0;31m'                                                    # Shell color RED
GREEN='\033[0;32m'                                                  # Shell color GREEN
YELLOW='\033[0;93m'                                                 # Shell color YELLOW
NC='\033[0m'                                                        # Shell color reset

################################################################
# specify needed sudo commands for agent after installation
################################################################
SUDO_COMMANDS="/usr/bin/docker stats*, /usr/bin/docker ps*, /usr/sbin/service --status-all, /usr/sbin/ufw status, /usr/sbin/iptables -S"

################################################################
# output Functions
################################################################
# print newline to stdout
print_newline() {
    printf "\n"
}

# print in green to stdout
print_green() {
    printf "${GREEN}${1}${NC}\n"
}

# print in white to stdout
print_white() {
    printf "${NC}${1}\n"
}

# print in yellow without newline to stdout
print_prompt() {
    printf "${YELLOW}${1}: ${NC}"
}

# print in yellow to stdout
print_yellow() {
    printf "${YELLOW}${1}${NC}\n"
}

# print in red to stderr
print_error() {
    printf >&2 "${RED}${1}${NC}\n"
}

# print shell-wide separator
print_separator() {
    echo "-----------------------------------------------------------"
}

# print script help
print_help() {
    print_separator
    print_white "$APP_NAME installer $INSTALLER_VERSION"
    print_separator
    print_white "Calling this script without arguments will install $APP_NAME"
    print_white ""
    print_white "-h | --help        Display this help page"
    print_white "-v                 Increase verbosity"
    print_white "--update           Update existing installation"
    print_white "--uninstall        Completely remove $APP_NAME from this machine"
    print_separator
    print_white "Environment variables used for the values in $CONFIG_FOLDER/$CONFIG_FILE"
    print_white ""
    print_white '$LARASHED_APP_ID'
    print_white '$LARASHED_APP_KEY'
    print_white '$LARASHED_APP_ENV'
    print_white '$LARASHED_SOCKET_TYPE'
    print_white '$LARASHED_SOCKET_ADDRESS'
    print_white '$LARASHED_ADDITIONAL_ARGUMENTS'

}

################################################################
# conditional check functions
################################################################
# check if calling user is root
check_root() {
    if [ $(id -u) -ne 0 ]; then
        print_error "Please run this script as root!"
        return 1
    else
        return 0
    fi
}

# choose downloader
check_download_tool() {
    if command -v wget > /dev/null 2>&1; then
        CHOSEN_DOWNLOADER="wget"
    elif command -v curl > /dev/null 2>&1; then
        CHOSEN_DOWNLOADER="curl"
    else
        print_error "Neither curl or wget are present! Please install one of them to continue."
        return 1
    fi
}

# check if sudo is installed
check_sudo() {
    if !(command -v sudo > /dev/null 2>&1); then
        print_error "Please install sudo before proceeding with the installation of $APP_NAME"
        return 1
    fi
    return 0
}

################################################################
# download related functions
################################################################
# downloads a given url
download_url() {
    SOURCE_URL="$1"
    DEST_PATH="$2"
    HIDE_PROGRESS="$3"

    # make prints
    if $VERBOSE; then
        print_yellow "Downloading using $CHOSEN_DOWNLOADER from: $SOURCE_URL"
        print_yellow "to: $DEST_PATH"
    elif !($HIDE_PROGRESS); then
        print_white "Downloading $SOURCE_URL"
    fi

    # download using either curl or wget
    if [ "$CHOSEN_DOWNLOADER" = "curl" ] && !($HIDE_PROGRESS); then
        curl -SL -o "$DEST_PATH" "$SOURCE_URL"
    elif [ "$CHOSEN_DOWNLOADER" = "curl" ] && $HIDE_PROGRESS; then
        curl -SLs -o "$DEST_PATH" "$SOURCE_URL"
    elif [ "$CHOSEN_DOWNLOADER" = "wget" ] && !($HIDE_PROGRESS); then
        wget --show-progress --progress=bar:nocscroll -q -O "$DEST_PATH" "$SOURCE_URL"
    elif [ "$CHOSEN_DOWNLOADER" = "wget" ] && $HIDE_PROGRESS; then
        wget -q -O "$DEST_PATH" "$SOURCE_URL"
    else
        print_error "Internal script error."
        return 1
    fi

    # error handling
    if [ $? -ne 0 ]; then
        print_error "ERROR! Download of $DEST_PATH failed."
        return 1
    elif !($HIDE_PROGRESS) || $VERBOSE; then
        print_green "Download of $DEST_PATH complete."
        return 0
    fi
}

# get latest binary version info from github api
get_github_release_information() {
    LINKLIST_NAME="download.linklist"

    # mktemp
    TEMP_WORK_DIR="$(mktemp -d -t $APP_NAME_SLUG.XXXXXXXX)" || { print_error "Error during mktemp!"; return 1; }

    # build API-URL
    API_URL="$(echo $GITHUB_URL |
        sed 's~github\.com~api\.github\.com/repos~g')/releases/latest"

    # download release information and parse
    download_url "$API_URL" "$TEMP_WORK_DIR/$LINKLIST_NAME" "true" || { print_error "Error during release information download!"; return 1; }
    API_RESPONSE=$(cat "$TEMP_WORK_DIR/$LINKLIST_NAME")

    LATEST_LINUX_BINARY_URL=$(printf '%s\n' "$API_RESPONSE" |
        grep -P "browser_download_url.*$LINUX_BINARY_NAME[^.]+" |
        cut -d":" -f2,3 | tr -d \" | tr -d " " | tr -d ",")
    LATEST_LINUX_BINARY_HASH_URL=$(printf '%s\n' "$API_RESPONSE" |
        grep -P "browser_download_url.*$LINUX_BINARY_NAME\." |
        cut -d":" -f2,3 | tr -d \" | tr -d " " | tr -d "," | head -n1)
    LATEST_LINUX_BINARY_VERSION=$(printf '%s\n' "$API_RESPONSE" |
        grep "tag_name" |
        cut -d":" -f2,3 | tr -d \" | tr -d " " | tr -d ",")
}

# download checksums + binary and verify
download_and_check() {
    DOWNLOAD_URL="${1}"
    DOWNLOAD_CHECKSUM_URL="${2}"
    CHECKSUM_COMMAND=$(echo -n "$DOWNLOAD_CHECKSUM_URL" | rev | cut -d"." -f1 | rev)

    # download files
    download_url "$DOWNLOAD_URL" "$TEMP_WORK_DIR/$APP_NAME_SLUG" "true" || { print_error "Error during binary download!"; return 1; }
    download_url "$DOWNLOAD_CHECKSUM_URL" "$TEMP_WORK_DIR/$APP_NAME_SLUG.$CHECKSUM_COMMAND" "true" || { print_error "Error during cheksum download!"; return 1; }

    print_green "Binary and checksum file downloaded."

    # verify checksum
    CHECKSUM_DOWNLOAD=$($CHECKSUM_COMMAND "$TEMP_WORK_DIR/$APP_NAME_SLUG" | cut -d " " -f1)
    CHECKSUM_COMPARE=$(cat "$TEMP_WORK_DIR/$APP_NAME_SLUG.$CHECKSUM_COMMAND" | cut -d " " -f1)
    if $VERBOSE; then
        print_yellow "Checksum algorithm:             $CHECKSUM_COMMAND"
        print_yellow "Checksum of downloaded binary:  $CHECKSUM_DOWNLOAD"
        print_yellow "Expected Checksum:              $CHECKSUM_COMPARE"
    fi
    if [ "$CHECKSUM_DOWNLOAD" = "$CHECKSUM_COMPARE" ]; then
        print_green "Checksum verification complete."
    else
        print_error "Checksum verification failed!"
        print_error "Checksum algorithm:            $CHECKSUM_COMMAND"
        print_error "Checksum of downloaded binary: $CHECKSUM_DOWNLOAD"
        print_error "Expected Checksum:             $CHECKSUM_COMPARE"
    fi
}

# clean up downloaded files
clean_exit() {
    if [ -z $TEMP_WORK_DIR ]; then
        if $VERBOSE; then
            print_yellow "TEMP_WORK_DIR not initialized, no cleanup needed."
        fi
    else
        if $VERBOSE; then
            print_yellow "Cleaning up $TEMP_WORK_DIR"
        fi
        rm -rf $TEMP_WORK_DIR || true
    fi
}

################################################################
# runtime config generator functions
################################################################
# generates config file in /etc
generate_config() {
    DESTINATION_PATH="${1}"

    # if not existent, create folder
    mkdir -p "$CONFIG_FOLDER" || { print_error "Error creating config folder: $CONFIG_FOLDER"; return 1; }

    # if config not in env, set default values
    if [ -z "$LARASHED_APP_ID" ]; then LARASHED_APP_ID="xxxxx"; fi
    if [ -z "$LARASHED_APP_KEY" ]; then LARASHED_APP_KEY="xxxxx"; fi
    if [ -z "$LARASHED_APP_ENV" ]; then LARASHED_APP_ENV="production"; fi
    if [ -z "$LARASHED_SOCKET_TYPE" ]; then LARASHED_SOCKET_TYPE="tcp"; fi
    if [ -z "$LARASHED_SOCKET_ADDRESS" ]; then LARASHED_SOCKET_ADDRESS="127.0.0.1:33101"; fi
    if [ -z "$LARASHED_ADDITIONAL_ARGUMENTS" ]; then LARASHED_ADDITIONAL_ARGUMENTS=""; fi

    # generate config
    echo "APP_ID=$LARASHED_APP_ID" > "$DESTINATION_PATH" || { print_error "Error creating config file: $CONFIG_FOLDER/$CONFIG_FILE"; return 1; }
    echo "APP_KEY=$LARASHED_APP_KEY" >> "$DESTINATION_PATH"
    echo "APP_ENV=$LARASHED_APP_ENV" >> "$DESTINATION_PATH"
    echo "SOCKET_TYPE=$LARASHED_SOCKET_TYPE" >> "$DESTINATION_PATH"
    echo "SOCKET_ADDRESS=$LARASHED_SOCKET_ADDRESS" >> "$DESTINATION_PATH"
    echo "ADDITIONAL_ARGUMENTS=$LARASHED_ADDITIONAL_ARGUMENTS" >> "$DESTINATION_PATH"

    if $VERBOSE; then
        print_yellow "Successfully created config in $DESTINATION_PATH"
    fi
}

# generates systemd unit file
install_systemd_unit() {
    CONFIG_FILE_PATH=$CONFIG_FILE_PATH/$CONFIG_FILE

    # build run args
    RUN_ARGS="run --app-id=\"\${APP_ID}\""
    RUN_ARGS="$RUN_ARGS --app-key=\"\${APP_KEY}\""
    RUN_ARGS="$RUN_ARGS --app-env=\"\${APP_ENV}\""
    RUN_ARGS="$RUN_ARGS --socket-type=\"\${SOCKET_TYPE}\""
    RUN_ARGS="$RUN_ARGS --socket-address=\"\${SOCKET_ADDRESS}\""
    RUN_ARGS="$RUN_ARGS \"\${ADDITIONAL_ARGUMENTS}\""

    ## build systemd unit
    # build unit part
    echo "[Unit]" > "$SYSTEMD_UNIT_PATH" || { print_error "Error creating systemd unit file!" ; return 1; }
    echo "Description=$APP_NAME $LATEST_LINUX_BINARY_VERSION" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "Documentation=$GITHUB_URL" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "After=syslog.target network.target remote-fs.target nss-lookup.target" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "" >> "$SYSTEMD_UNIT_PATH" || return 1

    # build service part
    echo "[Service]" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "EnvironmentFile=$CONFIG_FOLDER/$CONFIG_FILE" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "ExecStart=$BINARY_DESTINATION $RUN_ARGS" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "RestartSec=$SYSTEMD_RESTART_TIME" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "TimeoutSec=$SYSTEMD_TIMEOUT" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "User=$UNIX_USERNAME" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "Group=$UNIX_USERNAME" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "" >> "$SYSTEMD_UNIT_PATH" || return 1

    # build install part
    echo "[Install]" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "WantedBy=multi-user.target" >> "$SYSTEMD_UNIT_PATH" || return 1
    echo "" >> "$SYSTEMD_UNIT_PATH" || return 1

    # bet permissions on unit file
    chown root:root "$SYSTEMD_UNIT_PATH" || { print_error "Error setting permissions on $SYSTEMD_UNIT_PATH" ; return 1; }
    chmod 0664 "$SYSTEMD_UNIT_PATH" || { print_error "Error setting permissions on $SYSTEMD_UNIT_PATH" ; return 1; }

    # reload configs
    systemctl daemon-reload || { print_error "Error during systemctl daemon-reload" ; return 1; }

    # check that service is present
    systemctl status "$APP_NAME_SLUG.service" > /dev/null 2>&1
    if [ $? -eq "4" ]; then
        print_error "Check using \`systemctl status $APP_NAME_SLUG.service\`  failed!"
        return 1
    fi

    # if not updating, enable service
    if [ $PERFORM_UPDATE -ne 1 ]; then
        systemctl enable --no-pager --plain "$APP_NAME_SLUG.service" > /dev/null 2>&1 || true
    fi
    print_green "Successfully installed systemd unit."
}

# removes the larashed sudo entry
remove_sudo_entry(){
    OLD_ENTRIES=$(grep "$UNIX_USERNAME" /etc/sudoers)
    sed -i "/^$UNIX_USERNAME/d" "/etc/sudoers" > /dev/null 2>&1 || { print_error "Error removing sudo-entry!"; return 1; }
    if $VERBOSE; then
        print_yellow "Removed from sudoers:\n$OLD_ENTRIES"
    fi
}

# creates the larashed sudo entry
add_sudo_entry(){
    # check that entry is not already present
    if grep "$UNIX_USERNAME" "/etc/sudoers" > /dev/null 2>&1 ; then
        remove_sudo_entry
    fi
    SUDO_LINE="$UNIX_USERNAME ALL=(ALL) NOPASSWD: $SUDO_COMMANDS"
    echo "$SUDO_LINE" >> "/etc/sudoers" || { print_error "Error creating sudo-entry!"; return 1; }
    if $VERBOSE; then
        print_yellow "Added to sudoers:\n$SUDO_LINE"
    fi
}

################################################################
# actual install & uninstall workflow functions
################################################################
# Installs the agent after everything is downloaded, takes the path
# to the downloaded binary as argument 1
install_agent() {
    BINARY_DOWNLOAD_PATH="${1}"

    # check if binary is already in place
    if [ -f "$BINARY_DESTINATION" ] && [ "$PERFORM_UPDATE" -ne 1 ]; then
        print_error "Binary already in place! If you want to overwrite, please run this script with the --update option."
        return 1
    elif [ -f "$BINARY_DESTINATION" ] && [ "$PERFORM_UPDATE" -eq 1 ]; then
        systemctl stop --no-pager --plain "$APP_NAME_SLUG.service" > /dev/null 2>&1 || { print_error "Error stopping $APP_NAME_SLUG.service" ; return 1; }
        print_green "Successfully stopped $APP_NAME_SLUG.service for update"
    fi

    # copy binary to destination and set permissions
    cp -f "$BINARY_DOWNLOAD_PATH" "$BINARY_DESTINATION" || { print_error "Error copying downloaded binary: $BINARY_DOWNLOAD_PATH"; return 1; }
    chown root:root "$BINARY_DESTINATION" || { print_error "Error setting permissions on binary: $BINARY_DESTINATION"; return 1; }
    chmod 0755 "$BINARY_DESTINATION" || { print_error "Error setting permissions on binary: $BINARY_DESTINATION"; return 1; }

    # check if user exists, else create
    if !(getent passwd "$UNIX_USERNAME" > /dev/null 2>&1); then
        /usr/sbin/useradd -M "$UNIX_USERNAME" || { print_error "Error creating agent user: $UNIX_USERNAME"; return 1; }
        if $VERBOSE; then
            USER_UID=$(id -u $UNIX_USERNAME) || { print_error "Error getting user id of created user $UNIX_USERNAME"; return 1; }
            print_yellow "Created user $UNIX_USERNAME with UID $USER_UID."
        fi
    fi

    # check if docker group exists and if not create it
    if !(grep -q docker /etc/group); then
         groupadd docker
         print_yellow "Created docker group."
    fi

    # add our user to the docker group
    if !(grep -q docker /etc/group | grep "$UNIX_USERNAME"); then
        /usr/sbin/usermod -aG docker $UNIX_USERNAME || { print_error "Error adding $UNIX_USERNAME to the docker group"; }
    fi

    print_separator

    # check if config-folder is already in place
    if [ -f "$CONFIG_FOLDER/$CONFIG_FILE" ]; then
        print_yellow "Config file already present! Keeping the current version."
    # else, write config
    else
        generate_config "$CONFIG_FOLDER/$CONFIG_FILE" || { print_error "Error generating config."; return 1; }
        print_green "Successfully generated config."
    fi

    # set permissions on Config-Folder
    chown -R "$UNIX_USERNAME:$UNIX_USERNAME" "$CONFIG_FOLDER" || { print_error "Error setting permissions." ; return 1; }
    chmod -R 0770 "$CONFIG_FOLDER" || { print_error "Error setting permissions." ; return 1; }

    add_sudo_entry  || { print_error "Error setting sudo-entry." ; return 1; }

    # setup unit file
    install_systemd_unit || { print_error "Error installing systemd unit" ; return 1; }

    # start systemd service
    systemctl start $APP_NAME_SLUG.service || { print_error "Failed to start service" ; return 1; }
    print_green "Successfully started $APP_NAME_SLUG.service."

    print_separator
    print_green "Agent is installed and running."
    print_green "Agent configuration is stored in $CONFIG_FOLDER/$CONFIG_FILE."
    print_newline
}

# uninstalls the agent completely
uninstall_agent() {
    CHANGES_MADE=0

    # if present: stop & remove systemd unit
    systemctl status --no-pager --plain "$APP_NAME_SLUG.service" > /dev/null 2>&1
    if [ "$?" -ne "4" ]; then
        systemctl stop --no-pager --plain "$APP_NAME_SLUG.service" > /dev/null 2>&1 || { print_error "Error stopping $APP_NAME_SLUG.service" ; return 1; }
        systemctl disable --no-pager --plain "$APP_NAME_SLUG.service" > /dev/null 2>&1 || true
        rm -f "$SYSTEMD_UNIT_PATH" || { print_error "Error deleting $SYSTEMD_UNIT_PATH" ; return 1; }
        systemctl daemon-reload || { print_error "Error during systemctl daemon-reload" ; return 1; }
        systemctl reset-failed || { print_error "Error during systemctl reset-failed" ; return 1; }
        print_green "OK: Stopped & removed systemd unit"
        CHANGES_MADE=1
    elif !(systemctl status "$APP_NAME_SLUG.service" > /dev/null 2>&1) && $VERBOSE; then
        print_yellow "WARN: Found no systemd unit to remove"
    fi

    # if for whatever reason the binary is still running, stop it
    if ps -auxwe | grep -v "grep" | grep "$BINARY_DESTINATION"; then
        PID=$(ps -auxwe | grep -v "grep" | grep "$BINARY_DESTINATION" | tr -s ' ' | cut -d' ' -f2)
        timeout "$SYSTEMD_TIMEOUT" kill -15 "$PID" || kill -9 "$PID"
        print_yellow "WARN: Killed still running binary $BINARY_DESTINATION with PID $PID."
        CHANGES_MADE=1
    elif !(ps -auxwe | grep -v "grep" | grep "$BINARY_DESTINATION") && $VERBOSE; then
        print_green "OK: All agent processes are stopped."
    fi

    # if present, remove binary
    if [ -f "$BINARY_DESTINATION" ]; then
        rm -f "$BINARY_DESTINATION" || { print_error "Error deleting $BINARY_DESTINATION" ; return 1; }
        print_green "OK: Removed binary from $BINARY_DESTINATION."
        CHANGES_MADE=1
    elif !([ -f "$BINARY_DESTINATION" ]) && $VERBOSE; then
        print_yellow "WARN: Found no binary to delete."
    fi

    # if present, remove config folder
    if [ -d "$CONFIG_FOLDER" ]; then
        rm -rf "$CONFIG_FOLDER" || { print_error "Error deleting $CONFIG_FOLDER" ; return 1; }
        print_green "OK: Removed $CONFIG_FOLDER."
        CHANGES_MADE=1
    elif !([ -d "$CONFIG_FOLDER" ]) && $VERBOSE; then
        print_yellow "WARN: Found no config-folder to delete."
    fi

    # check if user exists and remove
    if getent passwd "$UNIX_USERNAME" > /dev/null 2>&1; then
        /usr/sbin/userdel "$UNIX_USERNAME" || { print_error "Error deleting user $UNIX_USERNAME" ; return 1; }
        print_green "OK: Removed user $UNIX_USERNAME."
        CHANGES_MADE=1
    elif !(getent passwd "$UNIX_USERNAME") && $VERBOSE; then
        print_yellow "WARN: Found no user to delete."
    fi

    # remove sudo-entry
    remove_sudo_entry  || { print_error "Error setting sudo-entry." ; return 1; }

    # make final print
    print_separator
    if [ $CHANGES_MADE = "1" ]; then
        print_green "OK: Successfully removed all agent components."
    else
        print_error "WARN: Cannot uninstall! Apparently the agent is not installed."
    fi
    print_newline
}

################################################################
# main area - call to functions
################################################################
# check if root or exit
check_root || { clean_exit; exit 1; }
# check that sudo is installed or exit
check_sudo || { clean_exit; exit 1; }


# parse arguments
SCRIPT_ARGS=$(getopt -o vh -l "uninstall,update,help" -- "$@")
if [ $? -ne 0 ]; then
    print_error "Incorrect arguments supplied."
    exit 1
fi
eval set -- "$SCRIPT_ARGS"

while true; do
    case "$1" in
    -v)  VERBOSE="true"
        ;;
    --uninstall)  SCRIPT_MODE="uninstall"
        ;;
    --update)  SCRIPT_MODE="install" && PERFORM_UPDATE=1
        ;;
    --help) print_help && exit 0
        ;;
    -h) print_help && exit 0
        ;;
    --)
        shift
        break
        ;;
    esac
    shift
done

# start install or uninstall workflow
if [ "$SCRIPT_MODE" = "install" ]; then
    # choose between wget and curl, exit on error
    check_download_tool || { clean_exit; exit 2; }

    # get info of the latest version
    get_github_release_information || { clean_exit; exit 3; }

    # print header
    print_separator
    print_white "Installing $APP_NAME $LATEST_LINUX_BINARY_VERSION"
    print_separator

    # download the latest version
    download_and_check "$LATEST_LINUX_BINARY_URL" "$LATEST_LINUX_BINARY_HASH_URL" || {
        print_error "Download failed"
        clean_exit
        exit 4
    }

    # install downloaded version
    install_agent "$TEMP_WORK_DIR/$APP_NAME_SLUG" || { print_error "Error during installation."; clean_exit ; exit 5; }

# uninstall part
elif [ "$SCRIPT_MODE" = "uninstall" ]; then
    uninstall_agent || { print_error "Uninstall failed."; clean_exit; exit 99; }
fi

# END
clean_exit
