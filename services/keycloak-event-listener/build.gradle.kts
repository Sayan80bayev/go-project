plugins {
    id("java")
    id("com.github.johnrengelman.shadow") version "8.1.1"
}

group = "com.example"
version = "1.0.0"

java {
    toolchain {
        languageVersion.set(JavaLanguageVersion.of(17))
    }
}

repositories {
    mavenCentral()
}

dependencies {
    // --- Keycloak SPIs (do NOT bundle them in your jar) ---
    compileOnly("org.keycloak:keycloak-server-spi:24.0.3")
    compileOnly("org.keycloak:keycloak-server-spi-private:24.0.3")
    compileOnly("org.keycloak:keycloak-core:24.0.3")

    // --- Your runtime deps ---
    implementation("org.apache.kafka:kafka-clients:3.6.1")
    implementation("org.slf4j:slf4j-api:2.0.13")

    // Optional: lombok
    compileOnly("org.projectlombok:lombok:1.18.32")
    annotationProcessor("org.projectlombok:lombok:1.18.32")

    testImplementation(platform("org.junit:junit-bom:5.10.0"))
    testImplementation("org.junit.jupiter:junit-jupiter")
}

tasks.test { useJUnitPlatform() }

// produce a thin jar (recommended). Keycloak provides its own libs.
tasks.jar {
    manifest {
        attributes(
            "Implementation-Title" to "Keycloak Kafka Event Listener",
            "Implementation-Version" to version
        )
    }
}

tasks {
    shadowJar {
        archiveBaseName.set("keycloak-event-listener")
        archiveClassifier.set("") // without "-all"
        archiveVersion.set("1.0.0")
    }
}