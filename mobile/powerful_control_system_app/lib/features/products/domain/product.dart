class Product {
  const Product({required this.id, required this.name, required this.price, required this.stock});

  final int id;
  final String name;
  final num price;
  final num stock;

  factory Product.fromJson(Map<String, dynamic> json) => Product(
        id: int.tryParse(json['id']?.toString() ?? '') ?? 0,
        name: json['nombre']?.toString() ?? json['descripcion']?.toString() ?? 'Producto',
        price: num.tryParse(json['precio_venta']?.toString() ?? json['precio']?.toString() ?? '0') ?? 0,
        stock: num.tryParse(json['stock_actual']?.toString() ?? json['cantidad']?.toString() ?? '0') ?? 0,
      );
}
