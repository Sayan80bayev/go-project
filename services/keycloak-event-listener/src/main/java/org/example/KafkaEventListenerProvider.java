package org.example;

import org.apache.kafka.clients.producer.KafkaProducer;
import org.apache.kafka.clients.producer.ProducerRecord;
import org.keycloak.events.Event;
import org.keycloak.events.EventListenerProvider;
import org.keycloak.events.EventType;
import org.keycloak.events.admin.AdminEvent;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.UserModel;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Properties;

public class KafkaEventListenerProvider implements EventListenerProvider {

    private static final Logger log = LoggerFactory.getLogger(KafkaEventListenerProvider.class);

    private final KeycloakSession session;
    private final KafkaProducer<String, String> producer;
    private final String topic;

    public KafkaEventListenerProvider(KeycloakSession session, String bootstrapServers, String topic) {
        this.session = session;
        this.topic = topic;

        Properties props = new Properties();
        props.put("bootstrap.servers", bootstrapServers);
        props.put("key.serializer", "org.apache.kafka.common.serialization.StringSerializer");
        props.put("value.serializer", "org.apache.kafka.common.serialization.StringSerializer");

        this.producer = new KafkaProducer<>(props);
        log.info("KafkaEventListener initialized with topic={} bootstrapServers={}", topic, bootstrapServers);
    }

    @Override
    public void onEvent(Event event) {
        // Слушаем регистрацию или логин
        if (event.getType() == EventType.REGISTER) {
            try {
                String userId = event.getUserId();
                String email = null;
                String firstName = null;
                String lastName = null;

                if (userId != null) {
                    UserModel user = session.users().getUserById(session.getContext().getRealm(), userId);
                    if (user != null) {
                        email = user.getEmail();
                        firstName = user.getFirstName();
                        lastName = user.getLastName();

                        // если поля в Keycloak отключены, но пришли из IdP:
                        if (firstName == null) {
                            firstName = user.getFirstAttribute("given_name");
                        }
                        if (lastName == null) {
                            lastName = user.getFirstAttribute("family_name");
                        }
                    }
                }

                String payload = String.format(
                        "{\"type\":\"UserCreated\",\"data\":{" +
                                "\"user_id\":\"%s\"," +
                                "\"email\":\"%s\"," +
                                "\"firstname\":\"%s\"," +
                                "\"lastname\":\"%s\"" +
                                "}}",
                        userId, email, firstName, lastName
                );


                producer.send(new ProducerRecord<>(topic, userId, payload));
                log.debug("Published user event: {}", payload);
            } catch (Exception e) {
                log.error("Failed to send user event to Kafka", e);
            }
        }
    }

    @Override
    public void onEvent(AdminEvent adminEvent, boolean includeRepresentation) {
//        try {
//            String payload = String.format(
//                    "{\"kind\":\"ADMIN\",\"operation\":\"%s\",\"realmId\":\"%s\",\"resourceType\":\"%s\",\"resourcePath\":\"%s\",\"time\":%d}",
//                    adminEvent.getOperationType(), adminEvent.getRealmId(),
//                    adminEvent.getResourceTypeAsString(), adminEvent.getResourcePath(), adminEvent.getTime()
//            );
//
//            producer.send(new ProducerRecord<>(topic, null, payload));
//            log.debug("Published admin event: {}", payload);
//        } catch (Exception e) {
//            log.error("Failed to send admin event to Kafka", e);
//        }
    }

    @Override
    public void close() {
        try {
            producer.flush();
            producer.close();
        } catch (Exception ignored) {}
    }
}