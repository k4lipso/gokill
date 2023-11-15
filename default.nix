{ pkgs, currentVendorHash, ... }:

pkgs.buildGoModule rec {
  pname = "gokill";
  version = "1.0";
  vendorHash = currentVendorHash;
  src = ./.;

  buildInputs = [
    pkgs.olm
  ];

  postInstall = ''
    cp -r ./etc $out/ #for .deb packages
    '';
}


