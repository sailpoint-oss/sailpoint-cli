# Where should we write the GIF?
Output assets/img/vhs/Windows.gif

# Set up a 1200x600 terminal with 46px font.
Set FontSize 16
Set Width 1200
Set Height 600

Hide Sleep 1s Show

# Start ...

# Pause for dramatic effect...
Sleep 300ms

# Type a command in the terminal.
Type "go build -o 'C:\Program Files\sailpoint\sail.exe'"

# Pause for dramatic effect...
Sleep 800ms

# Run the command by pressing enter.
Enter 

# Pause for dramatic effect...
Sleep 300ms

# Type a command in the terminal.
Type `Set-ItemProperty -Path 'Registry::HKEY_LOCAL_MACHINE\System\CurrentControlSet\Control\Session Manager\Environment' -Name PATH -Value "$((Get-ItemProperty -Path 'Registry::HKEY_LOCAL_MACHINE\System\CurrentControlSet\Control\Session Manager\Environment' -Name PATH).path);C:\Program Files\sailpoint"`

# Pause for dramatic effect...
Sleep 800ms

# Run the command by pressing enter.
Enter 

# Pause for dramatic effect...
Sleep 500ms

# Type a command in the terminal.
Type "sail"

# Pause for dramatic effect...
Sleep 700ms

# Run the command by pressing enter.
Enter

# Admire the output for a bit.
Sleep 5s