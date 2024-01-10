{ pkgs, currentVendorHash, ... }:

pkgs.buildGoModule rec {
  pname = "gokill";
  version = "0.5";
  vendorHash = currentVendorHash;
  src = ./.;

  buildInputs = [
    pkgs.olm
  ];

  postInstall = ''
    cp -r ./etc $out/ #for .deb packages
    '';
}


