package org.example;

import org.keycloak.Config;
import org.keycloak.events.EventListenerProvider;
import org.keycloak.events.EventListenerProviderFactory;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.KeycloakSessionFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class KafkaEventListenerProviderFactory implements EventListenerProviderFactory {

    private static final Logger log = LoggerFactory.getLogger(KafkaEventListenerProviderFactory.class);

    private String bootstrapServers;
    private String topic;

    @Override
    public EventListenerProvider create(KeycloakSession session) {
        return new KafkaEventListenerProvider(session, bootstrapServers, topic);
    }

    @Override
    public void init(Config.Scope config) {
        // Read from OS environment (Docker Compose sets these automatically)
        this.bootstrapServers = System.getenv().getOrDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:29092");
        this.topic = System.getenv().getOrDefault("KAFKA_TOPIC", "user-events");

        log.info("KafkaEventListenerFactory initialized with bootstrapServers={} topic={}", bootstrapServers, topic);
    }

    @Override
    public void postInit(KeycloakSessionFactory factory) {}

    @Override
    public void close() {}

    @Override
    public String getId() {
        return "kafka-event-listener";
    }
}