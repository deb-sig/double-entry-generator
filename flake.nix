{
  description = "Rule-based double-entry bookkeeping importer (from Alipay/WeChat/Huobi to Beancount)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        pkgs = import nixpkgs {
          inherit system;
        };
        buildDeps = with pkgs; [git go_1_19 gnumake];
        devDeps = with pkgs;
          buildDeps
          ++ [
            golangci-lint
          ];
      in rec {
        # `nix develop`
        devShell = pkgs.mkShell {buildInputs = devDeps;};
      }
    );
}
