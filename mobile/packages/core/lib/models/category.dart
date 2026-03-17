import 'package:equatable/equatable.dart';

/// A service category in the Seva taxonomy.
///
/// Categories form a tree: top-level categories (parentId == null) contain
/// subcategories which can themselves contain leaf categories.
class Category extends Equatable {
  final String id;
  final String name;
  final String slug;
  final String? description;
  final String? iconUrl;
  final String? parentId;
  final int sortOrder;
  final bool isActive;
  final int providerCount;
  final List<Category> children;

  const Category({
    required this.id,
    required this.name,
    required this.slug,
    this.description,
    this.iconUrl,
    this.parentId,
    this.sortOrder = 0,
    this.isActive = true,
    this.providerCount = 0,
    this.children = const [],
  });

  factory Category.fromJson(Map<String, dynamic> json) {
    return Category(
      id: json['id'] as String,
      name: json['name'] as String,
      slug: json['slug'] as String,
      description: json['description'] as String?,
      iconUrl: json['icon_url'] as String?,
      parentId: json['parent_id'] as String?,
      sortOrder: json['sort_order'] as int? ?? 0,
      isActive: json['is_active'] as bool? ?? true,
      providerCount: json['provider_count'] as int? ?? 0,
      children: (json['children'] as List<dynamic>?)
              ?.map((e) => Category.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'slug': slug,
      'description': description,
      'icon_url': iconUrl,
      'parent_id': parentId,
      'sort_order': sortOrder,
      'is_active': isActive,
      'provider_count': providerCount,
      'children': children.map((c) => c.toJson()).toList(),
    };
  }

  /// Whether this is a top-level category.
  bool get isTopLevel => parentId == null;

  Category copyWith({
    String? id,
    String? name,
    String? slug,
    String? description,
    String? iconUrl,
    String? parentId,
    int? sortOrder,
    bool? isActive,
    int? providerCount,
    List<Category>? children,
  }) {
    return Category(
      id: id ?? this.id,
      name: name ?? this.name,
      slug: slug ?? this.slug,
      description: description ?? this.description,
      iconUrl: iconUrl ?? this.iconUrl,
      parentId: parentId ?? this.parentId,
      sortOrder: sortOrder ?? this.sortOrder,
      isActive: isActive ?? this.isActive,
      providerCount: providerCount ?? this.providerCount,
      children: children ?? this.children,
    );
  }

  @override
  List<Object?> get props => [id, name, slug, parentId, isActive];
}
