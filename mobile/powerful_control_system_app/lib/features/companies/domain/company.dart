class Company {
  const Company({required this.id, required this.name, required this.type, required this.status});

  final int id;
  final String name;
  final String type;
  final String status;

  factory Company.fromJson(Map<String, dynamic> json) => Company(
        id: int.tryParse(json['id']?.toString() ?? '') ?? 0,
        name: json['nombre']?.toString() ?? '',
        type: json['tipo']?.toString() ?? '',
        status: json['estado']?.toString() ?? '',
      );
}
