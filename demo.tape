# Umaru CLI Demo Tape for VHS
# Render with: vhs demo.tape

Output .github/assets/demo.gif

Set Shell "powershell"
Set FontSize 14
Set Width 1100
Set Height 620
Set Padding 20
Set WindowBar "Colorful"
Set Theme "Catppuccin Mocha"

# Setup PATH so 'umaru' works without ./
Hide
Type '$env:Path = "e:\Projeler\UmaruCLI;$env:Path"; clear'
Enter
Sleep 1s
Show

# 1. Show all 17 available starter templates
Type "umaru list"
Sleep 600ms
Enter
Sleep 3.5s

# 2. Clear terminal
Type "clear"
Enter
Sleep 600ms

# 3. Demonstrate project scaffolding with Addons (Dry-Run mode)
Type "umaru init my-api -t go-fiber --db postgres --auth jwt --redis --dry-run"
Sleep 600ms
Enter
Sleep 4s

# 4. Clear terminal
Type "clear"
Enter
Sleep 600ms

# 5. Show global persistent configuration table
Type "umaru config list"
Sleep 600ms
Enter
Sleep 3.5s
