{
  description = "kokin-kugire — add kugire data to kokinwakashu TEI XML sources";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools       # goimports etc.
            libxml2       # xmllint for validation
            nano          # interactive editor (review)
            tmux          # new-window for align-review prepare
            direnv
            ollama        # LLM backend for semantic kugire suggestion
          ];

          shellHook = ''
            export GOPATH="$PWD/.go"
            export PATH="$GOPATH/bin:$PATH"
            if ! pgrep -x ollama > /dev/null; then
              echo "Starting ollama..."
              ollama serve &>/tmp/ollama.log &
            fi
          '';
        };
      });
}
