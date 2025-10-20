# First-Run CLI Tool

The CLI first-run tool provides a terminal-based credential capture flow.

## Steps
1. Run the executable from a shell. It creates `config/credentials.sqlite` if the
   file does not exist.
2. Enter a client display name when prompted.
3. Opt in to each service you want to configure and provide the requested
   values.
4. The tool performs simple validation (ensuring required fields are present)
   and reports success once credentials are saved.

This workflow is ideal for remote servers or automated provisioning scripts
where a GUI is not available.
