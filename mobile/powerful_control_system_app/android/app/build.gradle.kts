import java.util.Properties

plugins {
    id("com.android.application")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

val releaseKeystore = rootProject.file("key.properties")
val releaseKeystoreProperties = Properties().apply {
    if (releaseKeystore.exists()) {
        releaseKeystore.inputStream().use(::load)
    }
}

android {
    namespace = "com.powerfulcontrolsystem.pcs"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        applicationId = "com.powerfulcontrolsystem.pcs"
        // You can update the following values to match your application needs.
        // For more information, see: https://flutter.dev/to/review-gradle-config.
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    buildTypes {
        release {
            if (releaseKeystore.exists()) {
                signingConfig = signingConfigs.create("release") {
                    keyAlias = releaseKeystoreProperties["keyAlias"] as String
                    keyPassword = releaseKeystoreProperties["keyPassword"] as String
                    storeFile = file(releaseKeystoreProperties["storeFile"] as String)
                    storePassword = releaseKeystoreProperties["storePassword"] as String
                }
            }
        }
    }
}

gradle.taskGraph.whenReady {
    val releaseTaskRequested = allTasks.any { task ->
        task.name.equals("assembleRelease", ignoreCase = true) || task.name.equals("bundleRelease", ignoreCase = true)
    }
    if (releaseTaskRequested && !releaseKeystore.exists()) {
        throw GradleException("Falta android/key.properties para firmar una compilacion de distribucion.")
    }
}

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

flutter {
    source = "../.."
}
