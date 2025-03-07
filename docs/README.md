# HttpProbe Documentation

This directory contains the source files for the HttpProbe documentation website.

## Local Development

To run the documentation site locally:

1. Install Ruby and Bundler
2. Install dependencies:
   ```
   bundle install
   ```
3. Start the local server:
   ```
   bundle exec jekyll serve
   ```
4. Open your browser at `http://localhost:4000`

## Documentation Structure

- `index.md` - Home page
- `test-definitions.md` - Information about test definition files
- `variable-interpolation.md` - Details on variable interpolation
- `assertions.md` - Documentation for all assertion types
- `cli-usage.md` - Command line usage instructions
- `failure-reporting.md` - Understanding test failure reports
- `examples.md` - Example test definitions for various scenarios

## Contributing to Documentation

1. Fork the repository
2. Create a new branch for your changes
3. Make your edits
4. Open a pull request

## GitHub Pages

The documentation is automatically published to GitHub Pages when changes are pushed to the main branch.