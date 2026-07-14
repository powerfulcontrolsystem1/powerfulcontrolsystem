class MobileAccount {
  const MobileAccount({required this.email, required this.name, required this.role});

  final String email;
  final String name;
  final String role;

  factory MobileAccount.fromJson(Map<String, dynamic> json) => MobileAccount(
        email: json['email']?.toString() ?? '',
        name: json['name']?.toString() ?? '',
        role: json['role']?.toString() ?? '',
      );
}
