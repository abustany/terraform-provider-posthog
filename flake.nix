{
  description = "PostHog Terraform provider";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs";
    flake-utils.url  = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];
        pkgs = import nixpkgs {
          inherit system overlays;
        };
        rev = if (self ? shortRev) then self.shortRev else "dev";
      in
      with pkgs;
      {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.golangci-lint
            pkgs.gopls
            pkgs.terraform
          ];
        };

        packages.default = pkgs.buildGo120Module {
          pname = "terraform-provider-posthog";
          version = rev;
          src = pkgs.lib.cleanSource self;
          vendorHash = "sha256-LWbdzdqOtSXanOcwnQb1o16oBhNsR4ddn///+OdXMow=";
          postInstall = ''
            INSTALL_DIR=$out/hashicorp.com/abustany/posthog/0.0.1/$(go env GOOS)_$(go env GOARCH)
            mkdir -p $INSTALL_DIR
            mv $out/bin/terraform-provider-posthog $INSTALL_DIR/
            rmdir $out/bin
          '';
        };
      }
    );
}
