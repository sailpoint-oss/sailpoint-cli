package va

const TroubleshootingScript = "/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/luke-hagar-sp/VA-Scripts/main/stunt.sh)\""
const UpdateCommand = "sudo update_engine_client -check_for_update"
const RebootCommand = "sudo reboot"
