{ pkgs, lib, self, ... }:

with lib;
let
  docbuilder = self.packages.x86_64-linux.gokill-docbuilder;

  prepareMD = ''
    # Copy inputs into the build directory
    cp -r --no-preserve=all $inputs/* ./
    cp ${../README.md} ./README.md

    ${docbuilder}/bin/docbuilder --output ./
    substituteInPlace ./SUMMARY.md \
    --replace "@GOKILL_OPTIONS@" "$(${docbuilder}/bin/docbuilder)"

    cat ./SUMMARY.md
  '';
in
pkgs.stdenv.mkDerivation {
  name = "gokill-docs";
  phases = [ "buildPhase" ];
  buildInputs = [ pkgs.mdbook ];

  inputs = sourceFilesBySuffices ./. [ ".md" ".toml" ];

  buildPhase = ''
    dest=$out/share/doc
    mkdir -p $dest
    ${prepareMD}
    mdbook build
    cp -r ./book/* $dest
  '';
}
