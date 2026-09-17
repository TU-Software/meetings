{
  description = "simple chat application for docker demo";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-26.05";
    go-overlay.url = "github:purpleclay/go-overlay";
  };

  outputs =
    { nixpkgs, go-overlay, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ go-overlay.overlays.default ];
      };
      go = pkgs.go-bin.fromGoMod ./backend/go.mod;
    in
    {
      formatter.x86_64-linux = pkgs.alejandra;
      devShells.x86_64-linux.default = pkgs.mkShell {
        buildInputs = [
          pkgs.nixd
          pkgs.nil

          go.withDefaultTools
          pkgs.jq

          pkgs.bun
        ];
      };
    };
}
