part of '../finora_pages.dart';

/// Metadata shown by the home navigation shell.
class NavItem {
  const NavItem(this.title, this.icon);

  final String title;
  final IconData icon;
}

/// Declarative definition of an editable resource field.
class FieldSpec {
  const FieldSpec(this.key, this.label, {this.initial = ''});

  final String key;
  final String label;
  final String initial;
}
